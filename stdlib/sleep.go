package stdlib

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterBuiltin("sleep", &object.Builtin{
		Fn: sleep,
	})
}

func sleep(scope object.Scope, args ...object.Object) (object.Operation, error) {
	sl := &Sleep{unit: "s"}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Integer:
			sl.amount = arg.Value
		case *object.String:
			d, err := Interpolate(scope.Env.Export(), arg.Value)
			if err != nil {
				return nil, fmt.Errorf(err.Error())
			}
			last := d[len(d)-1:]
			_, err = strconv.Atoi(last)
			if err == nil {
				d = d + "s"
			}
			num := d[:len(d)-1]
			sl.unit = d[len(d)-1:]
			a, err := strconv.Atoi(num)
			if err != nil {
				return nil, fmt.Errorf(err.Error())
			}
			sl.amount = int64(a)
		default:
			return nil, fmt.Errorf("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}
	return func() object.Object {
		err := sl.Invoke()
		if err != nil {
			return object.NewError(err.Error())
		}
		return &object.Integer{Value: 0}
	}, nil
}

// Sleep represents and performs a `sleep` invocation
type Sleep struct {
	unit   string
	amount int64
}

// Invoke actually performs the sleep
func (sleep *Sleep) Invoke() error {
	var unitDur time.Duration
	switch sleep.unit {
	case "d":
		unitDur = time.Hour * 24
	case "s":
		unitDur = time.Second
	case "m":
		unitDur = time.Minute
	case "h":
		unitDur = time.Hour
	default:
		return errors.New("Invalid time interval " + sleep.unit)
	}
	time.Sleep(time.Duration(sleep.amount) * unitDur)
	return nil
}
