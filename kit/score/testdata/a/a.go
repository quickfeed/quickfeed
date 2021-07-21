package a

import (
	"testing"
)

func TestNoArgs()                          {}
func NotATest(*testing.T)                  {}
func TestFire(*testing.T)                  {}
func TestNoTParam(string)                  {}
func TesttooManyParams(*testing.T, string) {}
func TesttooManyNames(a, b *testing.T)     {}
func TestnoTParam(string)                  {}
