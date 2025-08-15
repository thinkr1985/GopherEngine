package main

import (
	"GopherEngine/core"
	"GopherEngine/guis"
	"GopherEngine/nomath"
	"math"
	"math/rand/v2"
	v2 "math/rand/v2"

	// "GopherEngine/nomath"
	_ "log"
	// "math/rand"
)

func populate_scene(scene *core.Scene) {
	scene.LoadAsset("E:/GitHub/GopherEngine/tests/GroundPlane.asset")

	i := 0
	for i < 100 {
		i += 1
		asmb := scene.LoadAsset("E:/GitHub/GopherEngine/tests/CartoonyTreeB.asset")
		asmb.Transform.SetPosition(nomath.Vec3{X: float64(v2.IntN(200)), Y: 0, Z: float64(v2.IntN(200))})
		scale_random := float64(v2.IntN(2))
		random := 0.1 + rand.Float64()*(10.0-0.1)
		asmb.Transform.SetScale(nomath.Vec3{X: scale_random, Y: scale_random, Z: scale_random})
		asmb.Transform.SetRotation(nomath.Vec3{X: 0, Y: random, Z: 0})
	}

	i = 0
	for i < 200 {
		i += 1
		asmb := scene.LoadAsset("E:/GitHub/GopherEngine/tests/CartoonyTreeA.asset")
		asmb.Transform.SetPosition(nomath.Vec3{X: float64(v2.IntN(400)), Y: 0, Z: float64(v2.IntN(400))})
		scale_random := float64(v2.IntN(10))
		random := 0.1 + rand.Float64()*(10.0-0.1)
		asmb.Transform.SetScale(nomath.Vec3{X: scale_random, Y: scale_random, Z: scale_random})
		asmb.Transform.SetRotation(nomath.Vec3{X: 0, Y: random, Z: 0})
	}

	i = 0
	for i < 25 {
		i += 1
		asmb := scene.LoadAsset("E:/GitHub/GopherEngine/tests/watchTower.asset")
		asmb.Transform.SetPosition(nomath.Vec3{X: float64(v2.IntN(100)), Y: 0, Z: float64(v2.IntN(200))})
		scale_random := float64(v2.IntN(50))
		random := 0.1 + rand.Float64()*(10.0-0.1)
		asmb.Transform.SetScale(nomath.Vec3{X: scale_random, Y: scale_random, Z: scale_random})
		asmb.Transform.SetRotation(nomath.Vec3{X: 0, Y: random, Z: 0})
	}
}

func main() {
	// core.StartCPUProfile()
	scene := core.NewScene()
	populate_scene(scene)
	scene.DefaultLight.Transform.Rotation.X += 3.0 + math.Sin(10)*1.0
	scene.DefaultLight.Transform.Dirty = true
	guis.Window(scene)
	// core.StopCPUProfile()
}
