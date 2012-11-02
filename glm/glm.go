package glm

import (
	"math"
)

type Vec3 struct{ Elem [3]float64 }
type Vec4 struct{ Elem [4]float64 }

type Mat3 struct{ Elem [3]Vec3 }
type Mat4 struct{ Elem [4] Vec4 }

func Inverse_sqrt(f float64) float64 {
	return 1.0 / math.Sqrt(f)
}

func NewVec3(x,y,z float64) *Vec3 {
	return &Vec3{[3]float64{x,y,z}}
}

func (v *Vec3) Copy() *Vec3 {
	b := *v
	return &b
}

func (v *Vec3) Scale(f float64) *Vec3 {
	r := v.Copy()
	r.Iscale(f)
	return r
}

func (v *Vec3) Iscale(f float64) *Vec3 {
	v.Elem[0] *= f
	v.Elem[1] *= f
	v.Elem[2] *= f
	return v
}

func (v *Vec3) Add(in *Vec3) *Vec3 {
	r := v.Copy()
	r.Iadd(in)
	return r
}

func (v *Vec3) Iadd(in *Vec3) *Vec3 {
	v.Elem[0] += in.Elem[0]
	v.Elem[1] += in.Elem[1]
	v.Elem[2] += in.Elem[2]
	return v
}

func (v *Vec3) Subtract(in *Vec3) *Vec3 {
	r := v.Copy()
	return r.Isubtract(in)
}

func (v *Vec3) Isubtract(in *Vec3) *Vec3 {
	v.Elem[0] -= in.Elem[0]
	v.Elem[1] -= in.Elem[1]
	v.Elem[2] -= in.Elem[2]
	return v
}

func (v *Vec3) Normalize() {
	sqr := v.Elem[0]*v.Elem[0] + v.Elem[1]*v.Elem[1] + v.Elem[2]*v.Elem[2]
	v.Iscale(Inverse_sqrt(sqr))
}

func (v *Vec3) Dot(in *Vec3) float64 {
	return v.Elem[0]*in.Elem[0] + v.Elem[1]*in.Elem[1] + v.Elem[2]*in.Elem[2]
}

func (v *Vec3) Cross(in *Vec3) *Vec3 {
	out := new(Vec3)
	out.Elem[0] = v.Elem[1]*in.Elem[2] - in.Elem[1]*v.Elem[2]
	out.Elem[1] = v.Elem[2]*in.Elem[0] - in.Elem[2]*v.Elem[0]
	out.Elem[2] = v.Elem[0]*in.Elem[1] - in.Elem[0]*v.Elem[1]
	return out
}

//-----------------------------------------------------------------------------

func NewVec4(x,y,z,w float64) *Vec4 {
	return &Vec4{[4]float64{x,y,z,w}}
}

func (v *Vec4) Copy() *Vec4 {
	b := *v
	return &b
}

func (v *Vec4) Scale(f float64) *Vec4 {
	b := v.Copy()
	b.Iscale(f)
	return b
}

func (v *Vec4) Iscale(f float64) *Vec4 {
	v.Elem[0] *= f
	v.Elem[1] *= f
	v.Elem[2] *= f
	v.Elem[3] *= f
	return v
}

func (v *Vec4) Add(in *Vec4) *Vec4 {
	r := v.Copy()
	r.Iadd(in)
	return r
}

func (v *Vec4) Iadd(in *Vec4) *Vec4 {
	v.Elem[0] += in.Elem[0]
	v.Elem[1] += in.Elem[1]
	v.Elem[2] += in.Elem[2]
	v.Elem[3] += in.Elem[3]
	return v
}

func (v *Vec4) Subtract(in *Vec4) *Vec4 {
	r := v.Copy()
	return r.Isubtract(in)
}

func (v *Vec4) Isubtract(in *Vec4) *Vec4 {
	v.Elem[0] -= in.Elem[0]
	v.Elem[1] -= in.Elem[1]
	v.Elem[2] -= in.Elem[2]
	v.Elem[3] -= in.Elem[3]
	return v
}

func (v *Vec4) Normalize() {
	sqr := v.Elem[0]*v.Elem[0] + v.Elem[1]*v.Elem[1] + v.Elem[2]*v.Elem[2] + v.Elem[3]*v.Elem[3]
	v.Iscale(Inverse_sqrt(sqr))
}

func (v *Vec4) Dot(in *Vec4) float64 {
	return v.Elem[0]*in.Elem[0] + v.Elem[1]*in.Elem[1] + v.Elem[2]*in.Elem[2] + v.Elem[3]*in.Elem[3]
}

//-----------------------------------------------------------------------------

func (m *Mat3) Copy() *Mat3 {
	b := *m
	return &b
}

func (m *Mat3) Scale(f float64) {
 m.Elem[0].Scale(f)
 m.Elem[1].Scale(f)
 m.Elem[2].Scale(f)
}

func (m *Mat3) Iadd(in *Mat3) {
 m.Elem[0].Iadd(&in.Elem[0])
 m.Elem[1].Iadd(&in.Elem[1])
 m.Elem[2].Iadd(&in.Elem[2])
}

func (m *Mat3) Mult(in *Vec3) *Vec3 {
	out := new(Vec3)
	out.Elem[0] = m.Elem[0].Dot(in)
	out.Elem[1] = m.Elem[1].Dot(in)
	out.Elem[2] = m.Elem[2].Dot(in)
	return out
}

func (m *Mat3) Multm(in *Mat3) *Mat3 {
	out := new(Mat3)
	t := in.Transpose()
	out.Elem[0] = Vec3{[3]float64{m.Elem[0].Dot(&t.Elem[0]), m.Elem[0].Dot(&t.Elem[1]), m.Elem[0].Dot(&t.Elem[2])}}
	out.Elem[1] = Vec3{[3]float64{m.Elem[1].Dot(&t.Elem[0]), m.Elem[1].Dot(&t.Elem[1]), m.Elem[1].Dot(&t.Elem[2])}}
	out.Elem[2] = Vec3{[3]float64{m.Elem[2].Dot(&t.Elem[0]), m.Elem[2].Dot(&t.Elem[1]), m.Elem[2].Dot(&t.Elem[2])}}
	return out
}

func (m *Mat3) Transpose() *Mat3 {
	result := new(Mat3)
	result.Elem[0].Elem[0] = m.Elem[0].Elem[0]
	result.Elem[0].Elem[1] = m.Elem[1].Elem[0]
	result.Elem[0].Elem[2] = m.Elem[2].Elem[0]

	result.Elem[1].Elem[0] = m.Elem[0].Elem[1]
	result.Elem[1].Elem[1] = m.Elem[1].Elem[1]
	result.Elem[1].Elem[2] = m.Elem[2].Elem[1]

	result.Elem[2].Elem[0] = m.Elem[0].Elem[2]
	result.Elem[2].Elem[1] = m.Elem[1].Elem[2]
	result.Elem[2].Elem[2] = m.Elem[2].Elem[2]
	return result
}

//-----------------------------------------------------------------------------

func (m *Mat4) Copy() *Mat4 {
	b := *m
	return &b
}

func (m *Mat4) Scale(f float64) {
 m.Elem[3].Scale(f)
 m.Elem[0].Scale(f)
 m.Elem[1].Scale(f)
 m.Elem[2].Scale(f)
}

func (m *Mat4) Add(in *Mat4) {
 m.Elem[3].Add(&in.Elem[3])
 m.Elem[0].Add(&in.Elem[0])
 m.Elem[1].Add(&in.Elem[1])
 m.Elem[2].Add(&in.Elem[2])
}

func (m *Mat4) Mult(in *Vec4) *Vec4 {
	out := new(Vec4)
	out.Elem[0] = m.Elem[0].Dot(in)
	out.Elem[1] = m.Elem[1].Dot(in)
	out.Elem[2] = m.Elem[2].Dot(in)
	out.Elem[3] = m.Elem[3].Dot(in)
	return out
}

func (m *Mat4) Multm(in *Mat4) *Mat4 {
	out := new(Mat4)
	t := in.Transpose()
	out.Elem[0] = Vec4{[4]float64{m.Elem[0].Dot(&t.Elem[0]),
	  m.Elem[0].Dot(&t.Elem[1]),
	  m.Elem[0].Dot(&t.Elem[2]),
	  m.Elem[0].Dot(&t.Elem[3])}}
	out.Elem[1] = Vec4{[4]float64{m.Elem[1].Dot(&t.Elem[0]),
	  m.Elem[1].Dot(&t.Elem[1]),
	  m.Elem[1].Dot(&t.Elem[2]),
	  m.Elem[1].Dot(&t.Elem[3])}}
	out.Elem[2] = Vec4{[4]float64{m.Elem[2].Dot(&t.Elem[0]),
    m.Elem[2].Dot(&t.Elem[1]),
    m.Elem[2].Dot(&t.Elem[2]),
    m.Elem[2].Dot(&t.Elem[3])}}
	out.Elem[3] = Vec4{[4]float64{m.Elem[3].Dot(&t.Elem[0]),
    m.Elem[3].Dot(&t.Elem[1]),
    m.Elem[3].Dot(&t.Elem[2]),
    m.Elem[3].Dot(&t.Elem[3])}}
	return out
}

func (m *Mat4) Transpose() *Mat4 {
	result := new(Mat4)
	result.Elem[0].Elem[0] = m.Elem[0].Elem[0]
	result.Elem[0].Elem[1] = m.Elem[1].Elem[0]
	result.Elem[0].Elem[2] = m.Elem[2].Elem[0]
	result.Elem[0].Elem[3] = m.Elem[3].Elem[0]

	result.Elem[1].Elem[0] = m.Elem[0].Elem[1]
	result.Elem[1].Elem[1] = m.Elem[1].Elem[1]
	result.Elem[1].Elem[2] = m.Elem[2].Elem[1]
	result.Elem[1].Elem[3] = m.Elem[3].Elem[1]

	result.Elem[2].Elem[0] = m.Elem[0].Elem[2]
	result.Elem[2].Elem[1] = m.Elem[1].Elem[2]
	result.Elem[2].Elem[2] = m.Elem[2].Elem[2]
	result.Elem[2].Elem[3] = m.Elem[3].Elem[2]

	result.Elem[3].Elem[0] = m.Elem[0].Elem[3]
	result.Elem[3].Elem[1] = m.Elem[1].Elem[3]
	result.Elem[3].Elem[2] = m.Elem[2].Elem[3]
	result.Elem[3].Elem[3] = m.Elem[3].Elem[3]
	return result
}
