package bot_test

import (
	. "github.com/natrim/grainbot/bot"
	"testing"
)

func TestNew(t *testing.T) {
	if NewBot() == nil {
		t.Error("cannot get new bot")
	}
}
