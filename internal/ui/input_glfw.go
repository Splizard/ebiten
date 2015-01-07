// Copyright 2015 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !js

package ui

import (
	glfw "github.com/go-gl/glfw3"
	"math"
)

func IsKeyPressed(key Key) bool {
	return current.input.isKeyPressed(key)
}

func IsMouseButtonPressed(button MouseButton) bool {
	return current.input.isMouseButtonPressed(button)
}

func CursorPosition() (x, y int) {
	return current.input.cursorPosition()
}

var glfwMouseButtonToMouseButton = map[glfw.MouseButton]MouseButton{
	glfw.MouseButtonLeft:   MouseButtonLeft,
	glfw.MouseButtonRight:  MouseButtonRight,
	glfw.MouseButtonMiddle: MouseButtonMiddle,
}

func (i *input) update(window *glfw.Window, scale int) {
	for g, e := range glfwKeyCodeToKey {
		i.keyPressed[e] = window.GetKey(g) == glfw.Press
	}
	for g, e := range glfwMouseButtonToMouseButton {
		i.mouseButtonPressed[e] = window.GetMouseButton(g) == glfw.Press
	}
	x, y := window.GetCursorPosition()
	i.cursorX = int(math.Floor(x)) / scale
	i.cursorY = int(math.Floor(y)) / scale
}