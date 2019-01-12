package wechat

import (
	"bytes"
	"encoding/xml"
	"io"

	"github.com/objcoding/wxpay"
)

// DecodeXML parses wxpay's weird response XML data.
// See https://github.com/objcoding/wxpay/issues/10
func DecodeXML(r io.Reader) wxpay.Params {
	var (
		d      *xml.Decoder
		start  *xml.StartElement
		params wxpay.Params
	)
	d = xml.NewDecoder(r)
	params = make(wxpay.Params)
	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			start = &t
		case xml.CharData:
			if t = bytes.TrimSpace(t); len(t) > 0 {
				params.SetString(start.Name.Local, string(t))
			}
		}
	}
	return params
}
