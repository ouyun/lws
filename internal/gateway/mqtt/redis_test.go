package mqtt

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gomodule/redigo/redis"
)

func TestNewRedisPool(t *testing.T) {
	pool := NewRedisPool()
	red := pool.Get()

	type intro struct {
		Title  string `redis:"Title"`
		Author string `redis:"Author"`
		Body   string `redis:"Body"`
	}

	var p1, p2, r1, r2 intro

	p1.Title = "Example"
	p1.Author = "Gary"
	p1.Body = "Hello"
	p2 = intro{
		Title:  "Example2",
		Author: "Steve",
		Body:   "Map",
	}

	if _, err := red.Do("HMSET", redis.Args{}.Add("id01").AddFlat(&p1)...); err != nil {
		panic(err)
	}

	m := map[string]string{
		"Title":  "Example2",
		"Author": "Steve",
		"Body":   "Map",
	}

	if _, err := red.Do("HMSET", redis.Args{}.Add("id02").AddFlat(m)...); err != nil {
		panic(err)
	}

	for _, id := range []string{"id01", "id02"} {

		v, err := redis.Values(red.Do("HGETALL", id))
		value, err := red.Do("HGETALL", id)
		fmt.Printf("value: %+s: \n", value)
		if err != nil {
			panic(err)
		}
		if id == "id01" {
			err := redis.ScanStruct(v, &r1)
			if err != nil {
				panic(err)
			}
		} else {
			err := redis.ScanStruct(v, &r2)
			if err != nil {
				panic(err)
			}
		}
	}
	fmt.Printf("p2: %+v\n", p2)
	fmt.Printf("r2: %+v\n", r2)
	if !reflect.DeepEqual(r1, p1) {
		t.Error("map p1 do not equal r1")
	}
	if !reflect.DeepEqual(r2, p2) {
		t.Error("map p2 do not equal r2")
	}
	defer red.Close()
}
