package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"runtime"
	"time"

	"gray/glm"
	"gray/scene"
)

const (
  MSAA = 4
	SUBPIXEL_OFFSET = float64(MSAA - 1) / 2.0
  JITTER = false
  REFLECTIONS = false
  MAX_DEPTH = 2
)

type response struct {
  y,x int
  colour glm.Vec3
}

func degree_to_rad(in float64) float64 {
	return math.Pi * in / 180.0
}

func background() *glm.Vec3 {
  return glm.NewVec3(0.1, 0.1, 0.1)
}

//TODO: write fail-early intersectShadowNodes
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


func trace(root []scene.Primitive, ambient glm.Vec3, ray, origin *glm.Vec3, lights []scene.Light, depth int) (*glm.Vec3, bool) {
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

	  // cast shadow ray.
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
        reflected_shadow_ray := shadow_ray.Subtract(normal.Scale(2*diffuse_coef))
        specular_coef := math.Abs(math.Pow(reflected_shadow_ray.Dot(ray), mat.Shininess))
        if specular_coef > 0.00001 {
          specular_tmp := mat.Specular.Scale(specular_coef)
          specular.Iadd(glm.NewVec3(specular_tmp.Elem[0]*light.Colour.Elem[0], specular_tmp.Elem[1]*light.Colour.Elem[1], specular_tmp.Elem[2]*light.Colour.Elem[2]))
        }
      }
    }
    colour.Iadd(&diffuse).Iadd(&specular)
    // cast reflectance ray.
    if REFLECTIONS {
      reflected := ray.Subtract(normal.Scale(2*normal.Dot(ray)))
      if (depth < MAX_DEPTH) {
        if reflected_color, hit := trace(root, ambient, reflected, intersection, lights, depth+1); hit {
          colour.Iscale(1 - mat.Mirror).Iadd(reflected_color.Scale(mat.Mirror))
        }
      }
    }
    return colour, true
  }
	return &glm.Vec3{}, false
}

func parallelTrace(output chan response, x,y int, root []scene.Primitive, ambient glm.Vec3, ray, origin *glm.Vec3, lights []scene.Light, depth int) {
  go func() {
    if ret, hit := trace(root, ambient, ray, origin, lights, depth); hit {
      output <- response{y,x, *ret}
    } else {
      output <- response{y,x, *background()}
    }
  }()
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
	pdone := 0
	// TODO: make less hobo
  // reserve unique channels for every ray that we want to trace
  pixels := scene.Height * scene.Width * MSAA * MSAA
	c := make(chan response, pixels)

  // Scatter rays in parallel
	for y := 0; y < scene.Height; y++ {
		for x := 0; x < scene.Width; x++ {
		  for yaa := 0; yaa < MSAA; yaa++ {
		    for xaa := 0; xaa < MSAA; xaa++ {
          x_offset := SUBPIXEL_OFFSET
          y_offset := SUBPIXEL_OFFSET
		      if JITTER {
		        x_offset += (rand.Float64() - 0.5) * (0.5/float64(MSAA))
		        y_offset += (rand.Float64() - 0.5) * (0.5/float64(MSAA))
          }
          subpixel := top_pixel.Add(hor.Scale(aspect_ratio * float64(x))).Add(
            scene.Up.Scale(float64(y))).Add(
            hor.Scale(aspect_ratio * (float64(xaa) - x_offset)/float64(MSAA))).Add(
            scene.Up.Scale((float64(yaa) - y_offset)/float64(MSAA)))
          ray := subpixel.Subtract(&scene.Eye)
          parallelTrace(c, x,y, scene.Primitives, scene.Ambient, ray, &scene.Eye, scene.Lights, 0)
        }
      }
    }
    if (100*y)/scene.Height - pdone >= 5 {
      pdone = (100*y)/scene.Height
      fmt.Println("Done ", pdone, "%...")
    }
  }
  fmt.Println("Done scatter, now gather")

  // Gather results
  acc := make([]glm.Vec3, scene.Height * scene.Width)
	for i := 0; i < pixels; i++ {
    result := <-c
    acc[result.y * scene.Width + result.x].Iadd(&result.colour)
  }
	for x := 0; x < scene.Width; x++ {
    for y := 0; y < scene.Height; y++ {
      a := acc[y * scene.Width + x]
      a.Iscale(1/float64(MSAA*MSAA))
      //clamp colour values.
      a.Elem[0] = math.Max(0.0, math.Min(a.Elem[0], 1.0))
      a.Elem[1] = math.Max(0.0, math.Min(a.Elem[1], 1.0))
      a.Elem[2] = math.Max(0.0, math.Min(a.Elem[2], 1.0))
      // pack into a color.RGBA struct
      pixel := color.RGBA{uint8(255*a.Elem[0]), uint8(255*a.Elem[1]), uint8(255*a.Elem[2]), 255}
      //fmt.Println(c)
      img_rect.Set(x, scene.Height - y - 1, pixel)
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
  t := time.Now()
  // Set up parallelism
  fmt.Println(runtime.NumCPU())
  runtime.GOMAXPROCS(runtime.NumCPU())
  fmt.Println("Starting tracing at ", t, "...")
  Render(scene)
  fmt.Println("Done tracing at ", time.Now())
  fmt.Println("Tracing took:", time.Now().Sub(t))
}
