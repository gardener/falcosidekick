package outputs

import (
	"bytes"
	"log"

	"github.com/falcosecurity/falcosidekick/types"
)

type header struct {
	Title    string `json:"title"`
	SubTitle string `json:"subtitle"`
}

type keyValue struct {
	TopLabel string `json:"topLabel"`
	Content  string `json:"content"`
}

type widget struct {
	KeyValue keyValue `json:"keyValue,omitempty"`
}

type section struct {
	Widgets []widget `json:"widgets"`
}

type card struct {
	Header   header    `json:"header,omitempty"`
	Sections []section `json:"sections,omitempty"`
}

type googlechatPayload struct {
	Text  string `json:"text,omitempty"`
	Cards []card `json:"cards,omitempty"`
}

func newGooglechatPayload(falcopayload types.FalcoPayload, config *types.Configuration) googlechatPayload {
	var messageText string
	widgets := []widget{}

	if config.Googlechat.MessageFormatTemplate != nil {
		buf := &bytes.Buffer{}
		if err := config.Googlechat.MessageFormatTemplate.Execute(buf, falcopayload); err != nil {
			log.Printf("[ERROR] : Error expanding Google Chat message %v", err)
		} else {
			messageText = buf.String()
		}
	}

	if config.Googlechat.OutputFormat == Text {
		return googlechatPayload{
			Text: messageText,
		}
	}

	for i, j := range falcopayload.OutputFields {
		var w widget
		switch v := j.(type) {
		case string:
			w = widget{
				KeyValue: keyValue{
					TopLabel: i,
					Content:  v,
				},
			}
		default:
			continue
		}

		widgets = append(widgets, w)
	}

	widgets = append(widgets, widget{KeyValue: keyValue{"rule", falcopayload.Rule}})
	widgets = append(widgets, widget{KeyValue: keyValue{"priority", falcopayload.Priority}})
	widgets = append(widgets, widget{KeyValue: keyValue{"time", falcopayload.Time.String()}})

	return googlechatPayload{
		Text: messageText,
		Cards: []card{
			{
				Sections: []section{
					{Widgets: widgets},
				},
			},
		},
	}
}

// GooglechatPost posts event to Google Chat
func (c *Client) GooglechatPost(falcopayload types.FalcoPayload) {
	err := c.Post(newGooglechatPayload(falcopayload, c.Config))
	if err != nil {
		c.Stats.GoogleChat.Add(Error, 1)
		c.PromStats.Outputs.With(map[string]string{"destination": "googlechat", "status": Error}).Inc()
	} else {
		c.Stats.GoogleChat.Add(OK, 1)
		c.PromStats.Outputs.With(map[string]string{"destination": "googlechat", "status": OK}).Inc()
	}

	c.Stats.GoogleChat.Add(Total, 1)
}