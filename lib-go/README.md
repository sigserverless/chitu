# funchan

Sharing states between serverless functions with differentiable data types. 

# example

Read `main.go`

Run: 

```shell 
go119 run main.go
```

# developing 

## develop a new data type 

For example, we develop a new data type `DDict`:

Basically, define the struct and `GetDType` method. 

```go
type DDict struct {
	DBase
}

func (*DDict) GetDType() DType {
	return DDICT
}
```

Add the `DType` enum in `dobj.go`. 

```go
type DType int

const (
	DVEC DType = iota
	DINT
	DDICT
)
```

Then define the `DVal` which holds the actual value of the data type. 

```go
type DDictVal struct {
	val map[string]any
}

func NewDDictVal() DVal {
	return &DDictVal{
		val: map[string]any{},
	}
}

func (*DDictVal) GetDType() DType {
	return DDICT
}

func (v *DDictVal) Apply(op WriteOp) {
	switch op.(type) {
	case *NonOp:
		// do nothing
		// Now we have no other WriteOp for this type
	default:
		panic(fmt.Sprintf("unsupported apply operation on DDict: %s", op.Kind()))
	}
}
```

Method `Await` associates the two types. 
```go
func (v *DDict) Await() map[string]any {
	return v.Proxy.DAwait(v.Id).(*DDictVal).val
}
```

Do not forget to add the mappings in `dobj.go`: 

```go
var DVAL_CONS = utils.MappingMaker(
	func(f func() DVal) DType { return f().GetDType() },
	[]func() DVal{..., NewDDictVal},
)

var DOBJ_CONS = func(base DBase) map[DType]DObj {
	return utils.MappingMaker(
		func(dobj DObj) DType { return dobj.GetDType() },
		[]DObj{..., &DDict{base}},
	)
}
```

Be patient! We are almost done. Add 

## develop a new read operation 

For example, we develop Read Operation `Length` on `DVec`:

```go 
type LengthOp struct{}

func (op *LengthOp) GetRetDType() DType {
	return DINT
}

func (v *DVec) Length() *DInt {
	rop := &LengthOp{}
	id := v.Proxy.DRead(v.Id, rop)
	return &DInt{
		DBase{
			Id:    id,
			Proxy: v.Proxy,
		},
	}
}
```

Add a case for all Write Operation on DVec:

```go
func (wop *PushOp) Diff(rop ReadOp) WriteOp {
	switch rop.(type) {
        ...
	case *LengthOp:
		return &PlusOp{1}
	default:
		panic("Some read operation on Push not implemented. ")
	}
}
```

## develop a more complicated write operation

For example, we develop Read Operation `Map` on `DVec`:

```go
type MapOp struct {
	F func(any) any
}

func (op *MapOp) GetRetDType() DType {
	return DVEC
}

func (v *DVec) Map(f func(any) any) *DVec {
	rop := &MapOp{
		F: f,
	}
	id := v.Proxy.DRead(v.Id, rop)
	return &DVec{
		DBase{
			Id:    id,
			Proxy: v.Proxy,
		},
	}
}
```

Then add `MapOp` differential on all Read operations: 

```go
func (wop *PushOp) Diff(rop ReadOp) WriteOp {
	switch rop := rop.(type) {
	...
	case *MapOp:
		return &PushOp{
			Elem: rop.F(wop.Elem),
		}
	default:
		panic("Some read operation on Push not implemented. ")
	}
}
```

## develop a new write operation

For example, we develop Write Operation `Push` on `DVec`:

```go
type PushOp struct {
	Elem any
}

func (*PushOp) Kind() string {
	return "Push"
}

func (v *DVec) Push(e any) {
	wop := &PushOp{
		Elem: e,
	}
	v.Proxy.DWrite(v.Id, wop)
}

func (wop *PushOp) Diff(rop ReadOp, rank int) WriteOp {
	switch rop := rop.(type) {
	case *LengthOp:
		return &PlusOp{1}
	case *MapOp:
		return &PushOp{
			Elem: rop.F(wop.Elem),
		}
	default:
		panic("Some read operation on Push not implemented. ")
	}
}
```

Add `Apply` case: 

```go
func (v *DVecVal) Apply(op WriteOp) {
	switch op := op.(type) {
	...
	case *PushOp:
		v.val = append(v.val, op.Elem)
	default:
		panic(fmt.Sprintf("unsupported apply operation on DVec: %s", op.Kind()))
	}
}
```

Add mapping: 

```go
var WriteOpMapping = utils.MappingMaker(
	func(wop WriteOp) string { return wop.Kind() },
	[]WriteOp{..., &PlusOp{}},
)
```