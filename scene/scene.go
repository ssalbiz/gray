package scene

import (
	"fmt"
	"math"

	"gray/glm"
)

const (
  Epsilon = 0.00001
  ONLY_DRAW_BOUNDS = false
)

type Light struct {
	Pos      glm.Vec3
	Colour glm.Vec3
	Falloff  glm.Vec3
}

type Material struct {
  Ambient glm.Vec3
  Diffuse glm.Vec3
  Specular glm.Vec3
  Shininess float64
}

type Primitive interface {
  GetMaterial() Material
  Intersect(ray, origin glm.Vec3) (b bool, retlen float64, normal glm.Vec3)
}

type Scene struct {
  Lights []Light
  Primitives []Primitive
  Eye, View, Up, Ambient glm.Vec3
  Width, Height int
  FOV float64
}

type Sphere struct {
  Pos glm.Vec3
  Rad float64
  Mat Material
}

type Box struct {
  Pos glm.Vec3 // bottom left
  Rad float64
  Mat Material
}

type Mesh struct {
  Verts []glm.Vec3
  Faces [][3]int
  Normals []glm.Vec3
  Bound Box
  Mat Material
}

/**
 * Real quadratic roots
 */
func quadraticRoots(A,B,C float64) []float64 {
  result := make([]float64, 2)
  if A == 0 {
    result[0] = -C/B
    return result[:1]
  }
  if t := B*B - 4*A*C; t > 0 {
    tt := math.Sqrt(t)
    //TODO(ssalbiz): make less dickish.
    result[0] = (-B + tt)/(2*A)
    result[1] = (-B - tt)/(2*A)
    return result
  }
  return nil
}

// MESH PRIMITIVES

func NewMesh(verts [][3]float64, faces [][3]int, mat Material) *Mesh {
  vec_verts := make([]glm.Vec3, len(verts))
  m := &Mesh{}
  min := [3]float64{verts[0][0], verts[0][1], verts[0][2]}
  max := [3]float64{verts[0][0], verts[0][1], verts[0][2]}

  for i, v := range verts {
    for j := 0; j < 3; j++ {
      min[j] = math.Min(min[j], v[j])
      max[j] = math.Max(max[j], v[j])
    }
    vec_verts[i] = *glm.NewVec3(v[0], v[1], v[2])
  }
  m.Verts = vec_verts
  m.Faces = faces
  // Generate plane normals for all faces.
  m.Normals = make([]glm.Vec3, len(faces))
  for i, f := range faces {
    v1 := vec_verts[f[0]].Subtract(&vec_verts[f[1]])
    v2 := vec_verts[f[0]].Subtract(&vec_verts[f[2]])
    m.Normals[i] = *v1.Cross(v2)
  }
  m.Bound.Pos = *glm.NewVec3(min[0], min[1], min[2])
  diff := glm.NewVec3(max[0], max[1], max[2]).Subtract(&m.Bound.Pos)
  m.Bound.Rad = math.Max(math.Max(diff.Elem[0], diff.Elem[1]), diff.Elem[2])
  m.Bound.Mat = mat
  m.Mat = mat
  return m
}

func major_axis(v glm.Vec3) (ret int) {
  if v.Elem[0] > v.Elem[1] && v.Elem[0] > v.Elem[2] {
    ret = 0
  } else if v.Elem[1] > v.Elem[0] && v.Elem[1] > v.Elem[2] {
    ret = 1
  } else {
    ret = 2
  }
  return
}

func drop_axis(v glm.Vec3, axis int) (ret glm.Vec3) {
  switch axis {
    case 0:
      ret = *glm.NewVec3(v.Elem[1], v.Elem[2], 0.0)
    case 1:
      ret = *glm.NewVec3(v.Elem[0], v.Elem[2], 0.0)
    case 2:
      ret = *glm.NewVec3(v.Elem[0], v.Elem[1], 0.0)
  }
  return
}

func project_2dface(face [3]int, verts []glm.Vec3, axis int) [3]glm.Vec3 {
  v1 := verts[face[0]]
  v2 := verts[face[1]]
  v3 := verts[face[2]]

  return [3]glm.Vec3{
    drop_axis(v1, axis),
    drop_axis(v2, axis),
    drop_axis(v3, axis) }
}

func same_side(a, b, c, test glm.Vec3) bool {
  t := b.Subtract(&a)
  return t.Cross(test.Subtract(&a)).Dot(t.Cross(c.Subtract(&a))) > Epsilon
}

func (p Mesh) Intersect(ray, origin glm.Vec3) (b bool, raylen float64, normal glm.Vec3) {
  // If we're debugging just draw the bounding box.
  if ONLY_DRAW_BOUNDS {
    return p.Bound.Intersect(ray, origin)
  }
  // Check bounding box.
  if b, _, _ := p.Bound.Intersect(ray, origin); !b {
    return false, 0, glm.Vec3{}
  }
  // Go through faces and do face intersections.
  raylen = 10000000.0
  for i, f := range p.Faces {
    ray_proj := p.Normals[i].Dot(&ray) // Intersect the ray with the plane.
    if math.Abs(ray_proj) > Epsilon {
      new_raylen := p.Verts[f[0]].Subtract(&origin).Dot(&p.Normals[i]) / ray_proj
      // check that the ray origin is not coincident with the plane
      if new_raylen > Epsilon {
        // project face to 2D
        maj_axis := major_axis(p.Normals[i])
        verts2d := project_2dface(f, p.Verts, maj_axis)
        test_pt := drop_axis(*origin.Add(ray.Scale(new_raylen)), maj_axis)
        // clip the ray to the bounds of the 2D face.
        if same_side(verts2d[0], verts2d[1], verts2d[2], test_pt) &&
           same_side(verts2d[1], verts2d[2], verts2d[0], test_pt) &&
           same_side(verts2d[2], verts2d[0], verts2d[1], test_pt) {
           b = true
           if new_raylen < raylen {
             raylen = new_raylen
             normal = p.Normals[i]
           }
        }
      }
    }
  }
  return
}

func (p Mesh) GetMaterial() Material {
  return p.Mat
}


// BOX PRIMITIVES

func (p Box) Intersect(ray, origin glm.Vec3) (b bool, raylen float64, normal glm.Vec3) {
  min := p.Pos
  max := p.Pos.Add(glm.NewVec3(p.Rad, p.Rad, p.Rad))
  raylen_near := -100000.0
  raylen_far :=   100000.0
  // Assume parallel intersections are not a thing.
  for i, raydir := range ray.Elem {
    if math.Abs(raydir) < Epsilon && (origin.Elem[i] < min.Elem[i] || origin.Elem[i] > max.Elem[i]) {
      return false, 0, glm.Vec3{}
    }
    t1 := (min.Elem[i] - origin.Elem[i]) / raydir
    t2 := (max.Elem[i] - origin.Elem[i]) / raydir
    flip := false
    if t1 > t2 {
      t := t1; t1 = t2; t2 = t; flip = true
    }
    if t1 > raylen_near {
      raylen_near = t1
      normal = *glm.NewVec3(0.0, 0.0, 0.0)
      if flip {
        normal.Elem[i] = p.Rad
      } else {
        normal.Elem[i] = -p.Rad
      }
    }
    raylen_far = math.Min(raylen_far, t2)
    if raylen_far < raylen_near || raylen_far < Epsilon {
      return false, 0, normal
    }
  }
  return true, raylen_near, normal
}

func (p Box) GetMaterial() Material {
  return p.Mat
}

// SPHERE PRIMITIVES

func ray_epsilon_check(raylen float64, ray glm.Vec3, line glm.Vec3) (b bool, retlen float64, normal glm.Vec3) {
  if raylen > Epsilon {
    return true, raylen, *ray.Scale(raylen).Subtract(&line)
  }
  //fmt.Println(raylen)
  return false, 0, *glm.NewVec3(0,0,0)
}

func (p Sphere) Intersect(ray, origin glm.Vec3) (b bool, raylen float64, normal glm.Vec3) {
  normal = glm.Vec3{}
  line := *p.Pos.Subtract(&origin)
  // Solve using cosine law for scalar coefficient of ray.
  raylens := quadraticRoots(ray.Dot(&ray), 2*line.Dot(&ray), line.Dot(&line) - p.Rad*p.Rad)
  switch len(raylens) {
    case 0: return false, 0, normal
    case 1:
      return ray_epsilon_check(-raylens[0], ray, line)
    case 2:
      if raylens[0] > raylens[1] { // NOTE: negative inversion
        tmp := raylens[0]; raylens[0] = raylens[1]; raylens[1] = tmp;
      }
      if b, raylen, normal = ray_epsilon_check(-raylens[1], ray, line); b {
        return b, raylen, normal
      } else {
        return ray_epsilon_check(-raylens[0], ray, line)
      }
  }
  return false, 0, normal
}

func (p Sphere) GetMaterial() Material {
  return p.Mat
}


// SCENE DESCRIPTION

func CreateScene() (scene *Scene, err error) {
  scene = &Scene{}
  scene.Lights = []Light{
    Light{*glm.NewVec3(-100.0, 150.0, 400.0), *glm.NewVec3(0.7, 0.7, 0.7), *glm.NewVec3(1.0, 0.0, 0.0)},
    Light{*glm.NewVec3( 400.0, 100.0, 150.0), *glm.NewVec3(0.7, 0.0, 0.7), *glm.NewVec3(1.0, 0.0, 0.0)},
  }
  mat1 := Material{
    *glm.NewVec3(0.7, 1.0, 0.7), *glm.NewVec3(0.7, 1.0, 0.7), *glm.NewVec3(0.5, 0.7, 0.5), 25.0,
  }
  mat2 := Material{
    *glm.NewVec3(0.5, 0.5, 0.5), *glm.NewVec3(0.5, 0.5, 0.5), *glm.NewVec3(0.5, 0.7, 0.5), 25.0,
  }
  mat3 := Material{
    *glm.NewVec3(1.0, 0.6, 0.1), *glm.NewVec3(1.0, 0.6, 0.1), *glm.NewVec3(0.5, 0.7, 0.5), 25.0,
  }
  mat4 := Material{
    *glm.NewVec3(0.7, 0.6, 1.0), *glm.NewVec3(0.7, 0.6, 1.0), *glm.NewVec3(0.5, 0.4, 0.8), 25.0,
  }

  scene.Primitives = make([]Primitive, 7)
  scene.Primitives[0] = Sphere{ *glm.NewVec3(0.0, 0.0, -400.0), 100.0, mat1 }
  scene.Primitives[1] = Sphere{ *glm.NewVec3(200.0, 50.0, -100.0), 150.0, mat1 }
  scene.Primitives[2] = Sphere{ *glm.NewVec3(0.0, -1200.0, -500.0), 1000.0, mat2 }
  scene.Primitives[3] = Sphere{ *glm.NewVec3(-100.0, 25.0, -300.0), 50.0, mat3 }
  scene.Primitives[4] = Sphere{ *glm.NewVec3(0.0, 100.0, -250.0), 25.0, mat1 }
  scene.Primitives[5] = Box{*glm.NewVec3(-200.0, -125.0, 0.0), 100, mat4 }
  steldodec.Mat = mat3
  scene.Primitives[6] = *steldodec

  scene.Eye = *glm.NewVec3(0.0, 0.0, 800.0)
  scene.View = *glm.NewVec3(0.0, 0.0, -1.0)
  scene.Up = *glm.NewVec3(0.0, 1.0, 0.0)
  scene.Ambient = *glm.NewVec3(0.3, 0.3, 0.3)
  scene.FOV = 50
  scene.Width = 512
  scene.Height = 512
  fmt.Println("Here we go!")
  return scene, nil
}
