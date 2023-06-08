# mini-gateway

#### 缝缝补补的玩具项目

##### 使用 http1.1 转 grpc 中间件，grpc 服务端需要注册 json 编解码器

```go
package codec

import (
	"encoding/json"
	"google.golang.org/protobuf/encoding/protojson"
	
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"
)

func init() {
	encoding.RegisterCodec(JSON{
		mso: protojson.MarshalOptions{AllowPartial: true},
		umo: protojson.UnmarshalOptions{AllowPartial: true},
	})
}

type JSON struct {
	mso protojson.MarshalOptions
	umo protojson.UnmarshalOptions
}

func (_ JSON) Name() string {
	return "json"
}

func (j JSON) Marshal(v interface{}) (out []byte, err error) {
	if pm, ok := v.(proto.Message); ok {
		b, err := j.mso.Marshal(pm)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return json.Marshal(v)
}

func (j JSON) Unmarshal(data []byte, v interface{}) (err error) {
	if pm, ok := v.(proto.Message); ok {
		return j.umo.Unmarshal(data, pm)
	}
	return json.Unmarshal(data, v)
}

```