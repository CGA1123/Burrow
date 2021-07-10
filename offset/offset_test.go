package offset_test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/linkedin/Burrow/expect"
	"github.com/linkedin/Burrow/offset"
)

type Message struct {
	Key       string
	Value     string
	Partition int32
	Offset    int64
}

func Test_Smoke(t *testing.T) {
	f, err := os.Open("./offsets.json")
	expect.Ok(t, err)

	var messages []Message
	expect.Ok(t, json.NewDecoder(f).Decode(&messages))

	for _, message := range messages {
		key, err := hex.DecodeString(message.Key)
		expect.Ok(t, err)

		val, err := hex.DecodeString(message.Value)
		expect.Ok(t, err)

		if len(val) == 0 {
			continue
		}

		_, err = offset.Decode(key, val)
		if err != nil && err != offset.ErrGroupMetadataMsg {
			expect.Ok(t, err)
		}
	}
}
