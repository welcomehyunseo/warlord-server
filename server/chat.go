package server

import "fmt"

type Chat struct {
	text          string  `json:"text"`
	bold          bool    `json:"bold"`
	italic        bool    `json:"italic"`
	underlined    bool    `json:"underlined"`
	strikethrough bool    `json:"strikethrough"`
	obfuscated    bool    `json:"obfuscated"`
	font          string  `json:"font"`
	color         string  `json:"color"`
	insertion     string  `json:"insertion"`
	extra         []*Chat `json:"extra"`
}

func NewChat(
	text string,
	bold bool,
	italic bool,
	underlined bool,
	strikethrough bool,
	obfuscated bool,
	font string,
	color string,
	insertion string,
	extra []*Chat,
) *Chat {
	return &Chat{
		text:          text,
		bold:          bold,
		italic:        italic,
		underlined:    underlined,
		strikethrough: strikethrough,
		obfuscated:    obfuscated,
		font:          font,
		color:         color,
		insertion:     insertion,
		extra:         extra,
	}
}

func (c *Chat) GetText() string {
	return c.text
}

func (c *Chat) GetBold() bool {
	return c.bold
}

func (c *Chat) GetItalic() bool {
	return c.italic
}

func (c *Chat) GetUnderlined() bool {
	return c.underlined
}

func (c *Chat) GetStrikethrough() bool {
	return c.strikethrough
}

func (c *Chat) GetObfuscated() bool {
	return c.obfuscated
}

func (c *Chat) GetFont() string {
	return c.font
}

func (c *Chat) GetColor() string {
	return c.color
}

func (c *Chat) GetInsertion() string {
	return c.insertion
}

func (c *Chat) GetExtra() []*Chat {
	return c.extra
}

func (c *Chat) String() string {
	return fmt.Sprintf(
		"{ text: %s, bold: %v, italic: %v, underlined: %v, strikethrough: %v, obfuscated: %v,"+
			"font: %s, color: %s, insertion: %s, extra: %+v }",
		c.text,
		c.bold,
		c.italic,
		c.underlined,
		c.strikethrough,
		c.obfuscated,
		c.font,
		c.color,
		c.insertion,
		c.extra,
	)
}
