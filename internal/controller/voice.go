package controller

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"mango/internal/service"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var SessionProcessor = make(map[string]SessionChannels)

type VoiceHandler struct {
	userService    service.UserService
	audioProcessor *AudioProcessor
}

func NewVoiceHandler(userService service.UserService) *VoiceHandler {
	audioProcessor := &AudioProcessor{
		connections: SessionProcessor,
	}
	return &VoiceHandler{userService: userService, audioProcessor: audioProcessor}
}

// Register 注册路由
func (v *VoiceHandler) Register(router *gin.RouterGroup) {
	userRouter := router.Group("/voice")
	{
		//userRouter.GET("/stream", v.stream)
		userRouter.GET("/stream", v.streamHandle)
		userRouter.GET("/session", v.HandleSSE)    // 建立SSE会话
		userRouter.POST("/upload", v.HandleUpload) // 上传音频

	}
}

type AudioProcessor struct {
	connections map[string]SessionChannels
	mutex       sync.Mutex
}

type SessionChannels struct {
	EventChan chan string // 带10个缓冲
	ErrorChan chan error  // 带1个缓冲（推荐更健壮）
}

func (v *VoiceHandler) HandleSSE(c *gin.Context) {
	// 设置SSE头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	//w.Header().Set("Access-Control-Allow-Origin", "*")

	// 生成唯一会话ID
	sessionID := fmt.Sprintf("sess_%d", time.Now().UnixNano())
	sessionChannel := SessionChannels{
		EventChan: make(chan string, 10), // 事件通道带缓冲
		ErrorChan: make(chan error, 1),   // 错误通道带1个缓冲（推荐）
	}

	// 注册连接
	v.audioProcessor.mutex.Lock()
	v.audioProcessor.connections[sessionID] = sessionChannel
	v.audioProcessor.mutex.Unlock()

	// 确保连接关闭时清理
	defer func() {
		v.audioProcessor.mutex.Lock()
		delete(v.audioProcessor.connections, sessionID)
		v.audioProcessor.mutex.Unlock()
		close(sessionChannel.ErrorChan)
		close(sessionChannel.EventChan)
	}()

	// 通知客户端连接已建立
	StreamSuccessResponse(c, map[string]interface{}{"session_id": sessionID})
	flusher, _ := c.Writer.(http.Flusher)
	flusher.Flush()

	// 保持连接活跃
	ticker := time.NewTicker(5 * time.Minute) // 保持连接心跳
	defer ticker.Stop()

	for {
		select {
		case test := <-sessionChannel.EventChan:
			StreamSuccessResponse(c, map[string]interface{}{"message": test})
			flusher.Flush()
		case err := <-sessionChannel.ErrorChan: // 错误通道（无论有无缓冲，select均会立即处理）
			StreamErrorResponse(c, err.Error())
			flusher.Flush()
			return
		case <-ticker.C:
			StreamCloseResponse(c, "timeout")
			flusher.Flush()
			return
		case <-c.Request.Context().Done():
			StreamCloseResponse(c, "disconnect")
			flusher.Flush()
			return
		}
	}
}

type HandleUploadRequest struct {
	SessionId string `json:"session_id" form:"session_id" binding:"required"`
	Audio     string `json:"audio" form:"audio" binding:"required"`
}

func (v *VoiceHandler) HandleUpload(c *gin.Context) {

	var request HandleUploadRequest
	err := c.ShouldBind(&request)
	if err != nil {
		fmt.Println("ShouldBind fail")
		c.JSON(http.StatusCreated, "ShouldBind fail")
		return
	}
	// 获取对应的事件通道
	v.audioProcessor.mutex.Lock()
	sessionChannel, ok := v.audioProcessor.connections[request.SessionId]
	v.audioProcessor.mutex.Unlock()
	if !ok {
		c.JSON(http.StatusCreated, "no valid X-Session-ID")
		return
	}

	rawData, err := base64.StdEncoding.DecodeString(request.Audio)
	if err != nil {
		fmt.Println("StdEncoding fail")
		sessionChannel.ErrorChan <- err
		return
	}
	result, err := executeOne(
		map[string]interface{}{
			"id":   1,
			"path": "",
		},
		map[string]interface{}{},
		rawData,
	)
	if err != nil || result.Code != 0 {
		fmt.Println("executeOne fail")
		sessionChannel.ErrorChan <- err
		return
	}
	exists, text := isTextExistsAndValid(*result)
	if exists {
		sessionChannel.EventChan <- text
		fmt.Println("convert success")
		c.JSON(http.StatusCreated, "convert success")
	} else {
		sessionChannel.ErrorChan <- fmt.Errorf("convert fail")
		fmt.Println("convert fail")
		c.JSON(http.StatusCreated, "convert fail")
	}
	return
}

// Protocol constants
const (
	PROTOCOL_VERSION    = 0b0001
	DEFAULT_HEADER_SIZE = 0b0001

	// Message Types
	FULL_CLIENT_REQUEST   = 0b0001
	AUDIO_ONLY_REQUEST    = 0b0010
	FULL_SERVER_RESPONSE  = 0b1001
	SERVER_ACK            = 0b1011
	SERVER_ERROR_RESPONSE = 0b1111

	// Message Type Specific Flags
	NO_SEQUENCE       = 0b0000
	POS_SEQUENCE      = 0b0001
	NEG_SEQUENCE      = 0b0010
	NEG_WITH_SEQUENCE = 0b0011

	// Message Serialization
	NO_SERIALIZATION   = 0b0000
	JSON_SERIALIZATION = 0b0001

	// Message Compression
	NO_COMPRESSION   = 0b0000
	GZIP_COMPRESSION = 0b0001
)

// WAV header structure
type WAVHeader struct {
	ChunkID       [4]byte
	ChunkSize     uint32
	Format        [4]byte
	Subchunk1ID   [4]byte
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Subchunk2ID   [4]byte
	Subchunk2Size uint32
}

// Response structure
type Response struct {
	IsLastPackage   bool        `json:"is_last_package"`
	PayloadSequence int32       `json:"payload_sequence,omitempty"`
	Seq             int32       `json:"seq,omitempty"`
	Code            uint32      `json:"code,omitempty"`
	PayloadMsg      interface{} `json:"payload_msg,omitempty"`
	PayloadSize     uint32      `json:"payload_size"`
}

// Request structures
type User struct {
	UID string `json:"uid"`
}

type Audio struct {
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
	Bits       int    `json:"bits"`
	Channel    int    `json:"channel"`
	Codec      string `json:"codec"`
}

type RequestConfig struct {
	ModelName  string `json:"model_name"`
	EnablePunc bool   `json:"enable_punc"`
}

type Request struct {
	User    User          `json:"user"`
	Audio   Audio         `json:"audio"`
	Request RequestConfig `json:"request"`
}

// ASR WebSocket Client
type AsrWsClient struct {
	AudioPath   string
	SuccessCode int
	SegDuration int
	WsURL       string
	UID         string
	Format      string
	Rate        int
	Bits        int
	Channel     int
	Codec       string
	AuthMethod  string
	HotWords    []string
	Streaming   bool
	Mp3SegSize  int
	ReqEvent    int
}

// NewAsrWsClient creates a new ASR WebSocket client
func NewAsrWsClient(audioPath string, options map[string]interface{}) *AsrWsClient {
	client := &AsrWsClient{
		AudioPath:   audioPath,
		SuccessCode: 1000,
		SegDuration: 200,
		WsURL:       "wss://openspeech.bytedance.com/api/v3/sauc/bigmodel_nostream",
		UID:         "test",
		Format:      "pcm",
		Rate:        16000,
		Bits:        16,
		Channel:     1,
		Codec:       "pcm",
		AuthMethod:  "none",
		Streaming:   true,
		Mp3SegSize:  1000,
		ReqEvent:    1,
	}

	// Apply options
	if val, ok := options["seg_duration"]; ok {
		if segDuration, ok := val.(int); ok {
			client.SegDuration = segDuration
		}
	}
	if val, ok := options["ws_url"]; ok {
		if wsURL, ok := val.(string); ok {
			client.WsURL = wsURL
		}
	}
	if val, ok := options["uid"]; ok {
		if uid, ok := val.(string); ok {
			client.UID = uid
		}
	}
	if val, ok := options["format"]; ok {
		if format, ok := val.(string); ok {
			client.Format = format
		}
	}
	if val, ok := options["codec"]; ok {
		if codec, ok := val.(string); ok {
			client.Codec = codec
		}
	}
	if val, ok := options["rate"]; ok {
		if rate, ok := val.(int); ok {
			client.Rate = rate
		}
	}
	if val, ok := options["bits"]; ok {
		if bits, ok := val.(int); ok {
			client.Bits = bits
		}
	}
	if val, ok := options["channel"]; ok {
		if channel, ok := val.(int); ok {
			client.Channel = channel
		}
	}
	if val, ok := options["streaming"]; ok {
		if streaming, ok := val.(bool); ok {
			client.Streaming = streaming
		}
	}

	return client
}

// generateHeader generates protocol header
func generateHeader(messageType, messageTypeSpecificFlags, serialMethod, compressionType byte, reservedData byte) []byte {
	header := make([]byte, 4)
	headerSize := byte(1)

	header[0] = (PROTOCOL_VERSION << 4) | headerSize
	header[1] = (messageType << 4) | messageTypeSpecificFlags
	header[2] = (serialMethod << 4) | compressionType
	header[3] = reservedData

	return header
}

// generateBeforePayload generates sequence bytes
func generateBeforePayload(sequence int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(sequence))
	return buf
}

// parseResponse parses server response
func parseResponse(res []byte) (*Response, error) {
	if len(res) < 4 {
		return nil, fmt.Errorf("response too short")
	}

	headerSize := res[0] & 0x0f
	messageType := res[1] >> 4
	messageTypeSpecificFlags := res[1] & 0x0f
	serializationMethod := res[2] >> 4
	messageCompression := res[2] & 0x0f

	headerExtensionsEnd := int(headerSize) * 4
	if len(res) < headerExtensionsEnd {
		return nil, fmt.Errorf("response header incomplete")
	}

	payload := res[headerExtensionsEnd:]
	result := &Response{
		IsLastPackage: false,
	}

	var payloadMsg []byte
	var payloadSize uint32

	if messageTypeSpecificFlags&0x01 != 0 {
		// receive frame with sequence
		if len(payload) < 4 {
			return nil, fmt.Errorf("payload too short for sequence")
		}
		seq := int32(binary.BigEndian.Uint32(payload[:4]))
		result.PayloadSequence = seq
		payload = payload[4:]
	}

	if messageTypeSpecificFlags&0x02 != 0 {
		// receive last package
		result.IsLastPackage = true
	}

	switch messageType {
	case FULL_SERVER_RESPONSE:
		if len(payload) < 4 {
			return nil, fmt.Errorf("payload too short for size")
		}
		payloadSize = binary.BigEndian.Uint32(payload[:4])
		payloadMsg = payload[4:]
	case SERVER_ACK:
		if len(payload) < 4 {
			return nil, fmt.Errorf("payload too short for ack")
		}
		seq := int32(binary.BigEndian.Uint32(payload[:4]))
		result.Seq = seq
		if len(payload) >= 8 {
			payloadSize = binary.BigEndian.Uint32(payload[4:8])
			payloadMsg = payload[8:]
		}
	case SERVER_ERROR_RESPONSE:
		if len(payload) < 8 {
			return nil, fmt.Errorf("payload too short for error")
		}
		code := binary.BigEndian.Uint32(payload[:4])
		result.Code = code
		payloadSize = binary.BigEndian.Uint32(payload[4:8])
		payloadMsg = payload[8:]
	}

	if payloadMsg != nil {
		if messageCompression == GZIP_COMPRESSION {
			reader, err := gzip.NewReader(bytes.NewReader(payloadMsg))
			if err != nil {
				return nil, fmt.Errorf("gzip decompression failed: %v", err)
			}
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			if err != nil {
				return nil, fmt.Errorf("gzip read failed: %v", err)
			}
			payloadMsg = decompressed
		}

		if serializationMethod == JSON_SERIALIZATION {
			var jsonData interface{}
			if err := json.Unmarshal(payloadMsg, &jsonData); err != nil {
				return nil, fmt.Errorf("json unmarshal failed: %v", err)
			}
			result.PayloadMsg = jsonData
		} else if serializationMethod != NO_SERIALIZATION {
			result.PayloadMsg = string(payloadMsg)
		}
	}

	result.PayloadSize = payloadSize
	return result, nil
}

// readWAVInfo reads WAV file information
func readWAVInfo(data []byte) (int, int, int, int, []byte, error) {
	if len(data) < 44 {
		return 0, 0, 0, 0, nil, fmt.Errorf("WAV file too short")
	}

	reader := bytes.NewReader(data)
	var header WAVHeader

	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return 0, 0, 0, 0, nil, fmt.Errorf("failed to read WAV header: %v", err)
	}

	if string(header.ChunkID[:]) != "RIFF" || string(header.Format[:]) != "WAVE" {
		return 0, 0, 0, 0, nil, fmt.Errorf("invalid WAV file format")
	}

	nchannels := int(header.NumChannels)
	sampwidth := int(header.BitsPerSample / 8)
	framerate := int(header.SampleRate)
	nframes := int(header.Subchunk2Size) / (nchannels * sampwidth)
	waveBytes := data[44:]

	return nchannels, sampwidth, framerate, nframes, waveBytes, nil
}

// isWAV checks if data is WAV format
func isWAV(data []byte) bool {
	if len(data) < 44 {
		return false
	}
	return string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE"
}

// constructRequest creates request payload
func (c *AsrWsClient) constructRequest(reqID string) *Request {
	return &Request{
		User: User{
			UID: c.UID,
		},
		Audio: Audio{
			Format:     c.Format,
			SampleRate: c.Rate,
			Bits:       c.Bits,
			Channel:    c.Channel,
			Codec:      c.Codec,
		},
		Request: RequestConfig{
			ModelName:  "bigmodel",
			EnablePunc: true,
		},
	}
}

// sliceData splits data into chunks
func sliceData(data []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	dataLen := len(data)

	for offset := 0; offset < dataLen; offset += chunkSize {
		end := offset + chunkSize
		if end > dataLen {
			end = dataLen
		}
		chunks = append(chunks, data[offset:end])
	}

	return chunks
}

// segmentDataProcessor processes audio data in segments
func (c *AsrWsClient) segmentDataProcessor(wavData []byte, segmentSize int) (*Response, error) {
	reqID := uuid.New().String()
	seq := int32(1)

	requestParams := c.constructRequest(reqID)
	payloadBytes, err := json.Marshal(requestParams)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(payloadBytes); err != nil {
		return nil, fmt.Errorf("gzip write failed: %v", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("gzip close failed: %v", err)
	}
	payloadBytes = buf.Bytes()

	// Create full client request
	fullClientRequest := generateHeader(FULL_CLIENT_REQUEST, POS_SEQUENCE, JSON_SERIALIZATION, GZIP_COMPRESSION, 0x00)
	fullClientRequest = append(fullClientRequest, generateBeforePayload(seq)...)

	// Add payload size
	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, uint32(len(payloadBytes)))
	fullClientRequest = append(fullClientRequest, sizeBytes...)
	fullClientRequest = append(fullClientRequest, payloadBytes...)

	// Setup headers
	headers := map[string][]string{
		"X-Api-Resource-Id": {"volc.bigasr.sauc.duration"},
		"X-Api-Access-Key":  {"Z_ta1KD7amn7JMQOSwZfaG2TTYh2hhga"},
		"X-Api-App-Key":     {"5391266979"},
		"X-Api-Request-Id":  {reqID},
	}

	// Connect to WebSocket
	dialer := websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
	}

	conn, resp, err := dialer.Dial(c.WsURL, headers)
	if err != nil {
		return nil, fmt.Errorf("websocket connection failed: %v", err)
	}
	defer conn.Close()

	fmt.Println("Response headers:", resp.Header)

	// Send initial request
	if err := conn.WriteMessage(websocket.BinaryMessage, fullClientRequest); err != nil {
		return nil, fmt.Errorf("failed to send initial request: %v", err)
	}

	// Read initial response
	_, message, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read initial response: %v", err)
	}

	result, err := parseResponse(message)
	if err != nil {
		return nil, fmt.Errorf("failed to parse initial response: %v", err)
	}

	fmt.Println("******************")
	fmt.Printf("sauc result: %+v\n", result)
	fmt.Println("******************")

	// Process audio chunks
	chunks := sliceData(wavData, segmentSize)
	for i, chunk := range chunks {
		seq++
		isLast := (i == len(chunks)-1)

		if isLast {
			seq = -seq
		}

		start := time.Now()

		// Compress chunk
		var chunkBuf bytes.Buffer
		chunkGz := gzip.NewWriter(&chunkBuf)
		if _, err := chunkGz.Write(chunk); err != nil {
			return nil, fmt.Errorf("chunk gzip write failed: %v", err)
		}
		if err := chunkGz.Close(); err != nil {
			return nil, fmt.Errorf("chunk gzip close failed: %v", err)
		}
		compressedChunk := chunkBuf.Bytes()

		// Create audio only request
		var audioOnlyRequest []byte
		if isLast {
			audioOnlyRequest = generateHeader(AUDIO_ONLY_REQUEST, NEG_WITH_SEQUENCE, JSON_SERIALIZATION, GZIP_COMPRESSION, 0x00)
		} else {
			audioOnlyRequest = generateHeader(AUDIO_ONLY_REQUEST, POS_SEQUENCE, JSON_SERIALIZATION, GZIP_COMPRESSION, 0x00)
		}

		audioOnlyRequest = append(audioOnlyRequest, generateBeforePayload(seq)...)

		// Add payload size
		chunkSizeBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(chunkSizeBytes, uint32(len(compressedChunk)))
		audioOnlyRequest = append(audioOnlyRequest, chunkSizeBytes...)
		audioOnlyRequest = append(audioOnlyRequest, compressedChunk...)

		// Send audio chunk
		if err := conn.WriteMessage(websocket.BinaryMessage, audioOnlyRequest); err != nil {
			return nil, fmt.Errorf("failed to send audio chunk: %v", err)
		}

		// Read response
		_, message, err := conn.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf("failed to read chunk response: %v", err)
		}

		result, err = parseResponse(message)
		if err != nil {
			return nil, fmt.Errorf("failed to parse chunk response: %v", err)
		}

		fmt.Printf("%s, seq %d res %+v\n", time.Now().Format("2006-01-02 15:04:05.000"), seq, result)

		if c.Streaming {
			elapsed := time.Since(start)
			sleepTime := time.Duration(float64(c.SegDuration)*float64(time.Millisecond)) - elapsed
			if sleepTime > 0 {
				time.Sleep(sleepTime)
			}
		}
	}

	return result, nil
}

// Execute runs the ASR process
func (c *AsrWsClient) Execute(data []byte) (*Response, error) {
	var audioData []byte
	if data != nil {
		audioData = data
	} else {
		data, err := os.ReadFile(c.AudioPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read audio file: %v", err)
		}
		audioData = data
	}
	var segmentSize int

	switch c.Format {
	case "mp3":
		segmentSize = c.Mp3SegSize
	case "wav":
		nchannels, sampwidth, framerate, _, wavData, err := readWAVInfo(audioData)
		if err != nil {
			return nil, fmt.Errorf("failed to read WAV info: %v", err)
		}
		audioData = wavData
		sizePerSec := nchannels * sampwidth * framerate
		segmentSize = int(float64(sizePerSec) * float64(c.SegDuration) / 1000.0)
	case "pcm":
		segmentSize = int(float64(c.Rate) * 2.0 * float64(c.Channel) * float64(c.SegDuration) / 500.0)
	default:
		return nil, fmt.Errorf("unsupported format: %s", c.Format)
	}

	return c.segmentDataProcessor(audioData, segmentSize)
}

// executeOne processes a single audio item
func executeOne(audioItem map[string]interface{}, options map[string]interface{}, data []byte) (*Response, error) {
	_, ok := audioItem["id"]
	if !ok {
		return nil, fmt.Errorf("error")
	}
	audioPath, ok := audioItem["path"].(string)
	if !ok {
		return nil, fmt.Errorf("error")
	}

	client := NewAsrWsClient(audioPath, options)
	result, err := client.Execute(data)
	return result, err
}

// testStream tests streaming ASR
func (v *VoiceHandler) stream(c *gin.Context) {
	fmt.Println("测试流式")
	result, error := executeOne(
		map[string]interface{}{
			"id":     1,
			"path":   "/Users/liangyawei/isheji/mango/cmd/server/output.wav",
			"codec":  "pcm",
			"format": "pcm",
		},
		map[string]interface{}{},
		nil,
	)
	if error != nil {

	}
	fmt.Printf("Result: %+v\n", result)
}

// VoiceRequest 定义请求结构
type VoiceRequest struct {
	AudioData  string `json:"audio_data"`
	Text       string `json:"text"`
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
	Channels   int    `json:"channels"`
	BitDepth   int    `json:"bit_depth"`
}

// VoiceResponse 定义响应结构（根据你的需求调整）
type VoiceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

type StreamResponse struct {
	ID     int    `json:"id"`
	Path   string `json:"path"`
	Result struct {
		IsLastPackage   bool `json:"is_last_package"`
		PayloadSequence int  `json:"payload_sequence"`
		PayloadMsg      struct {
			AudioInfo struct {
				Duration int `json:"duration"`
			} `json:"audio_info"`
			Result struct {
				Additions struct {
					LogID string `json:"log_id"`
				} `json:"additions"`
				Text string `json:"text"` // 明确声明 text 字段
			} `json:"result"`
		} `json:"payload_msg"`
		PayloadSize int `json:"payload_size"`
	} `json:"result"`
}

// handleVoiceStream 处理语音流请求
func (v *VoiceHandler) streamHandle(c *gin.Context) {
	result, error := executeOne(
		map[string]interface{}{
			"id":   1,
			"path": "/Users/liangyawei/isheji/mango/cmd/server/audioa.wav",
		},
		map[string]interface{}{},
		nil,
	)
	if error != nil || result.Code != 0 {
		c.JSON(http.StatusCreated, "executeOne error")
	}

	exists, text := isTextExistsAndValid(*result)
	if exists {
		c.JSON(http.StatusCreated, text)
		fmt.Println("text不存在或为空")
	} else {
		c.JSON(http.StatusCreated, "false")
	}

}

func StreamSuccessResponse(c *gin.Context, data interface{}) {
	c.SSEvent("message", data)
}

// StreamErrorResponse 流式响应
func StreamErrorResponse(c *gin.Context, msg interface{}) {
	c.SSEvent("error", msg)
}

func StreamCloseResponse(c *gin.Context, data interface{}) {
	c.SSEvent("close", data)
}

func isTextExistsAndValid(resp Response) (bool, string) {
	// 2. 断言 PayloadMsg 为 map[string]interface{}
	payloadMsg, ok := resp.PayloadMsg.(map[string]interface{})
	if !ok {
		return false, ""
	}

	// 3. 检查 result 字段
	result, ok := payloadMsg["result"].(map[string]interface{})
	if !ok {
		return false, ""
	}

	// 4. 检查 text 字段
	text, exists := result["text"]
	if !exists {
		return false, ""
	}

	// 5. 确保 text 是字符串且非空
	textStr, ok := text.(string)
	if !ok || textStr == "" {
		return false, ""
	}

	return true, textStr
}
