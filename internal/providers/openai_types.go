package providers

import "encoding/json"

// OpenAI API response types (internal)

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Usage   *openAIUsage   `json:"usage,omitempty"`
}

type openAIChoice struct {
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type openAIMessage struct {
	Role             string           `json:"role"`
	Content          string           `json:"content"`
	ReasoningContent string           `json:"reasoning_content,omitempty"`
	Reasoning        string           `json:"reasoning,omitempty"` // Ollama alias for reasoning_content
	ToolCalls        []openAIToolCall `json:"tool_calls,omitempty"`
}

type openAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function openAIFunctionCall `json:"function"`
}

type openAIFunctionCall struct {
	Name             string   `json:"name"`
	Arguments        flexArgs `json:"arguments"`                   // Chấp nhận cả string và object (Gemini compat)
	ThoughtSignature string   `json:"thought_signature,omitempty"` // Gemini 2.5/3: must echo back
}

// flexArgs là custom type cho tool call arguments.
// OpenAI trả arguments dưới dạng JSON string: "arguments": "{\"path\": \"...\"}"
// Nhưng Gemini qua OpenAI-compat có thể trả dạng object: "arguments": {"path": "..."}
// Type này xử lý cả hai trường hợp, tránh json.Unmarshal thất bại làm mất SSE chunk.
type flexArgs string

func (f *flexArgs) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*f = ""
		return nil
	}
	// Trường hợp chuẩn OpenAI: arguments là JSON string (bọc trong dấu nháy kép)
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*f = flexArgs(s)
		return nil
	}
	// Trường hợp Gemini: arguments là raw JSON object/array — giữ nguyên dạng string
	*f = flexArgs(data)
	return nil
}

type openAIUsage struct {
	PromptTokens            int                      `json:"prompt_tokens"`
	CompletionTokens        int                      `json:"completion_tokens"`
	TotalTokens             int                      `json:"total_tokens"`
	PromptTokensDetails     *openAIPromptDetails     `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails *openAICompletionDetails `json:"completion_tokens_details,omitempty"`
}

type openAIPromptDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type openAICompletionDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

// Streaming types

type openAIStreamChunk struct {
	Choices []openAIStreamChoice `json:"choices"`
	Usage   *openAIUsage         `json:"usage,omitempty"`
}

type openAIStreamChoice struct {
	Delta        openAIStreamDelta `json:"delta"`
	FinishReason string            `json:"finish_reason,omitempty"`
}

type openAIStreamDelta struct {
	Content          string                 `json:"content,omitempty"`
	ReasoningContent string                 `json:"reasoning_content,omitempty"`
	Reasoning        string                 `json:"reasoning,omitempty"` // Ollama alias for reasoning_content
	ToolCalls        []openAIStreamToolCall `json:"tool_calls,omitempty"`
}

type openAIStreamToolCall struct {
	Index    int                `json:"index"`
	ID       string             `json:"id,omitempty"`
	Function openAIFunctionCall `json:"function"`
}

// toolCallAccumulator extends ToolCall with temporary fields for accumulating
// streamed arguments and thought_signature during SSE streaming.
type toolCallAccumulator struct {
	ToolCall
	rawArgs    string
	thoughtSig string
}
