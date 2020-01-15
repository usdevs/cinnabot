package cinnabot

import (
	"testing"
	"time"
)

type printable interface {
	string() string
}

// Just prints output format when using -v flag. Doesn't actually test anything.
func TestStrings(t *testing.T) {

	// washer off
	w1 := washer{
		Name:               "A",
		Level:              17,
		Ezlink:             false,
		Washer:             true,
		On:                 false,
		TimeChanged:        time.Now().Add(time.Duration(time.Minute * -10)),
		TimeChangedCertain: true,
		LastSeen:           time.Now().Add(time.Duration(time.Minute * -2)),
	}

	// washer on, < 30 min
	w2 := washer{
		Name:               "A",
		Level:              17,
		Ezlink:             false,
		Washer:             true,
		On:                 true,
		TimeChanged:        time.Now().Add(time.Duration(time.Minute * -20)),
		TimeChangedCertain: true,
		LastSeen:           time.Now().Add(time.Duration(time.Minute * -2)),
	}

	// washer off, hasn't been seen in a while
	w3 := washer{
		Name:               "A",
		Level:              17,
		Ezlink:             false,
		Washer:             true,
		On:                 false,
		TimeChanged:        time.Now().Add(time.Duration(time.Minute * -10)),
		TimeChangedCertain: true,
		LastSeen:           time.Now().Add(time.Duration(time.Minute * -20)),
	}

	// washer on, uncertain start time
	w4 := washer{
		Name:               "A",
		Level:              17,
		Ezlink:             false,
		Washer:             true,
		On:                 true,
		TimeChanged:        time.Now().Add(time.Duration(time.Minute * -10)),
		TimeChangedCertain: false,
		LastSeen:           time.Now().Add(time.Duration(time.Minute * -2)),
	}

	// washer on, overran slightly (and uncertain start time)
	w5 := washer{
		Name:               "A",
		Level:              17,
		Ezlink:             false,
		Washer:             true,
		On:                 true,
		TimeChanged:        time.Now().Add(time.Duration(time.Minute * -32)),
		TimeChangedCertain: false,
		LastSeen:           time.Now().Add(time.Duration(time.Minute * -2)),
	}

	// washer on, overran alot (likely broken sensor)
	w6 := washer{
		Name:               "A",
		Level:              17,
		Ezlink:             false,
		Washer:             true,
		On:                 true,
		TimeChanged:        time.Now().Add(time.Duration(time.Minute * -39)),
		TimeChangedCertain: true,
		LastSeen:           time.Now().Add(time.Duration(time.Minute * -2)),
	}

	d1 := dryer{
		Name:               "A",
		Level:              17,
		Ezlink:             false,
		Washer:             false,
		On:                 false,
		TimeChanged:        time.Now().Add(time.Duration(time.Minute * -39)),
		TimeChangedCertain: true,
		LastSeen:           time.Now().Add(time.Duration(time.Minute * -2)),
	}

	d2 := dryer{
		Name:               "A",
		Level:              17,
		Ezlink:             false,
		Washer:             false,
		On:                 true,
		TimeChanged:        time.Now().Add(time.Duration(time.Minute * -20)),
		TimeChangedCertain: true,
		LastSeen:           time.Now().Add(time.Duration(time.Minute * -2)),
	}

	d3 := dryer{
		Name:               "A",
		Level:              17,
		Ezlink:             false,
		Washer:             false,
		On:                 true,
		TimeChanged:        time.Now().Add(time.Duration(time.Minute * -50)),
		TimeChangedCertain: true,
		LastSeen:           time.Now().Add(time.Duration(time.Minute * -2)),
	}

	machines := []printable{w1, w2, w3, w4, w5, w6, d1, d2, d3}
	for i, m := range machines {
		t.Log(i + 1)
		t.Log(m.string())
	}
}
