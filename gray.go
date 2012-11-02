package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"gray/glm"
	"gray/scene"
)

func degree_to_rad(in float64) float64 {
	return math.Pi * in / 180.0
}

func intersectNodes(nodes []scene.Primitive, ray, origin *glm.Vec3) (any bool, min_node int, min_raylen float64, min_normal glm.Vec3) {
  min_raylen = 10000.0
  for i, node := range nodes {
    if hit, raylen, normal := node.Intersect(*ray, *origin); hit && raylen < min_raylen {
      any = true
      min_node = i
      min_raylen = raylen
      min_normal = normal
    }
  }
  return
}


func trace(root []scene.Primitive, ambient glm.Vec3, ray, origin *glm.Vec3, lights []scene.Light) color.RGBA {
	if hit, node, raylen, normal := intersectNodes(root, ray, origin); hit {
	  // ambient silhouette
	  mat := root[node].GetMaterial()
	  colour := glm.NewVec3(ambient.Elem[0]*mat.Ambient.Elem[0], ambient.Elem[1]*mat.Ambient.Elem[1], ambient.Elem[2]*mat.Ambient.Elem[2])
	  // setup for casting secondary (shadow) rays.
	  intersection := origin.Add(ray.Scale(raylen))
	  diffuse := glm.Vec3{}
	  specular := glm.Vec3{}
    ray.Normalize()
    normal.Normalize()
	  // cast shadow rays.
	  for _, light := range lights {
      shadow_ray := light.Pos.Subtract(intersection)
      if hit, _, _, _ = intersectNodes(root, shadow_ray, intersection); !hit {
        shadow_ray.Normalize()
        // add diffuse/specular components.
        diffuse_coef := normal.Dot(shadow_ray)
        if diffuse_coef > 0.00001 {
          diffuse_tmp := mat.Diffuse.Scale(diffuse_coef)
          diffuse.Iadd(glm.NewVec3(diffuse_tmp.Elem[0]*light.Colour.Elem[0], diffuse_tmp.Elem[1]*light.Colour.Elem[1], diffuse_tmp.Elem[2]*light.Colour.Elem[2]))
        }
        reflected_ray := normal.Scale(2*diffuse_coef).Subtract(shadow_ray)
        specular_coef := math.Abs(math.Pow(reflected_ray.Dot(ray), mat.Shininess))
        if specular_coef > 0.00001 {
          specular_tmp := mat.Specular.Scale(specular_coef)
          specular.Iadd(glm.NewVec3(specular_tmp.Elem[0]*light.Colour.Elem[0], specular_tmp.Elem[1]*light.Colour.Elem[1], specular_tmp.Elem[2]*light.Colour.Elem[2]))
        }
      }
    }
    colour.Iadd(&diffuse).Iadd(&specular)
    //clamp colour values.
    colour.Elem[0] = math.Max(0.0, math.Min(colour.Elem[0], 1.0))
    colour.Elem[1] = math.Max(0.0, math.Min(colour.Elem[1], 1.0))
    colour.Elem[2] = math.Max(0.0, math.Min(colour.Elem[2], 1.0))

    return color.RGBA{uint8(255*colour.Elem[0]), uint8(255*colour.Elem[1]), uint8(255*colour.Elem[2]), 255}
  }
	return color.RGBA{10, 10, 10, 255}
}

func Render(scene *scene.Scene) {
	// hard coded scene.
	var aspect_ratio float64 = float64(scene.Width) / float64(scene.Height)
	var view_len float64 = (float64(scene.Height) / math.Tan(degree_to_rad(scene.FOV)/2.0)) / 2.0

	hor := scene.View.Cross(&scene.Up)
	top_pixel := scene.Eye.Copy().Iadd(scene.View.Scale(view_len))
	hor.Normalize()
	scene.Up.Normalize()
	top_pixel.Iadd(hor.Scale(float64(-scene.Width) / 2.0))
	top_pixel.Iadd(scene.Up.Scale(float64(-scene.Height) / 2.0))

	img_rect := image.NewRGBA(image.Rect(0, 0, scene.Width, scene.Height))

	// dump scene info
	fmt.Println("Scene info:")
	fmt.Println(scene)

	for y := 0; y < scene.Height; y++ {
		for x := 0; x < scene.Width; x++ {
			pixel := top_pixel.Add(hor.Scale(aspect_ratio * float64(x))).Add(
				scene.Up.Scale(float64(y)))
			ray := pixel.Subtract(&scene.Eye)
			c := trace(scene.Primitives, scene.Ambient, ray, &scene.Eye, scene.Lights)
			//fmt.Println(c)
			img_rect.Set(x, scene.Height - y - 1, c)
		}
	}

	w, err := os.Create("out.png")
	defer w.Close()
	png.Encode(w, img_rect)
	if (err != nil) {
	  fmt.Println(err)
	}
}

func main() {
  scene, err := scene.CreateScene()
  if err != nil {
    fmt.Println(err)
    fmt.Println()
    return
  }
  Render(scene)
}
