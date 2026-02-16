package scripts

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/redis/go-redis/v9"
)

//go:embed compare_and_update.lua
var compareAndUpdateLua string

//go:embed update_if_lowest.lua
var updateIfLowestLua string

var CompareAndUpdateScript = redis.NewScript(compareAndUpdateLua)
var UpdateIfLowestScript = redis.NewScript(updateIfLowestLua)

// UpdateIfGreater runs the compare_and_update Lua script with ">" operator.
func UpdateIfGreater(ctx context.Context, client redis.Scripter, key, field string, value float64) *redis.Cmd {
	return CompareAndUpdateScript.Run(ctx, client, []string{key}, field, fmt.Sprintf("%v", value), ">")
}

// UpdateIfLess runs the compare_and_update Lua script with "<" operator.
func UpdateIfLess(ctx context.Context, client redis.Scripter, key, field string, value float64) *redis.Cmd {
	return CompareAndUpdateScript.Run(ctx, client, []string{key}, field, fmt.Sprintf("%v", value), "<")
}
