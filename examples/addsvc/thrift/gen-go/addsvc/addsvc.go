// Autogenerated by Thrift Compiler (1.0.0-dev)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING

package addsvc

import (
	"bytes"
	"reflect"
	"context"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
)

// (needed to ensure safety because of naive import list construction.)
var _ = thrift.ZERO
var _ = fmt.Printf
var _ = context.Background
var _ = reflect.DeepEqual
var _ = bytes.Equal

// Attributes:
//  - Value
//  - Err
type SumReply struct {
  Value int64 `thrift:"value,1" db:"value" json:"value"`
  Err string `thrift:"err,2" db:"err" json:"err"`
}

func NewSumReply() *SumReply {
  return &SumReply{}
}


func (p *SumReply) GetValue() int64 {
  return p.Value
}

func (p *SumReply) GetErr() string {
  return p.Err
}
func (p *SumReply) Read(iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 1:
      if fieldTypeId == thrift.I64 {
        if err := p.ReadField1(iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(fieldTypeId); err != nil {
          return err
        }
      }
    case 2:
      if fieldTypeId == thrift.STRING {
        if err := p.ReadField2(iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *SumReply)  ReadField1(iprot thrift.TProtocol) error {
  if v, err := iprot.ReadI64(); err != nil {
  return thrift.PrependError("error reading field 1: ", err)
} else {
  p.Value = v
}
  return nil
}

func (p *SumReply)  ReadField2(iprot thrift.TProtocol) error {
  if v, err := iprot.ReadString(); err != nil {
  return thrift.PrependError("error reading field 2: ", err)
} else {
  p.Err = v
}
  return nil
}

func (p *SumReply) Write(oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin("SumReply"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField1(oprot); err != nil { return err }
    if err := p.writeField2(oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *SumReply) writeField1(oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin("value", thrift.I64, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:value: ", p), err) }
  if err := oprot.WriteI64(int64(p.Value)); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.value (1) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:value: ", p), err) }
  return err
}

func (p *SumReply) writeField2(oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin("err", thrift.STRING, 2); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:err: ", p), err) }
  if err := oprot.WriteString(string(p.Err)); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.err (2) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 2:err: ", p), err) }
  return err
}

func (p *SumReply) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("SumReply(%+v)", *p)
}

// Attributes:
//  - Value
//  - Err
type ConcatReply struct {
  Value string `thrift:"value,1" db:"value" json:"value"`
  Err string `thrift:"err,2" db:"err" json:"err"`
}

func NewConcatReply() *ConcatReply {
  return &ConcatReply{}
}


func (p *ConcatReply) GetValue() string {
  return p.Value
}

func (p *ConcatReply) GetErr() string {
  return p.Err
}
func (p *ConcatReply) Read(iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 1:
      if fieldTypeId == thrift.STRING {
        if err := p.ReadField1(iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(fieldTypeId); err != nil {
          return err
        }
      }
    case 2:
      if fieldTypeId == thrift.STRING {
        if err := p.ReadField2(iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *ConcatReply)  ReadField1(iprot thrift.TProtocol) error {
  if v, err := iprot.ReadString(); err != nil {
  return thrift.PrependError("error reading field 1: ", err)
} else {
  p.Value = v
}
  return nil
}

func (p *ConcatReply)  ReadField2(iprot thrift.TProtocol) error {
  if v, err := iprot.ReadString(); err != nil {
  return thrift.PrependError("error reading field 2: ", err)
} else {
  p.Err = v
}
  return nil
}

func (p *ConcatReply) Write(oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin("ConcatReply"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField1(oprot); err != nil { return err }
    if err := p.writeField2(oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *ConcatReply) writeField1(oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin("value", thrift.STRING, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:value: ", p), err) }
  if err := oprot.WriteString(string(p.Value)); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.value (1) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:value: ", p), err) }
  return err
}

func (p *ConcatReply) writeField2(oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin("err", thrift.STRING, 2); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:err: ", p), err) }
  if err := oprot.WriteString(string(p.Err)); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.err (2) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 2:err: ", p), err) }
  return err
}

func (p *ConcatReply) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("ConcatReply(%+v)", *p)
}

type AddService interface {
  // Parameters:
  //  - A
  //  - B
  Sum(ctx context.Context, a int64, b int64) (r *SumReply, err error)
  // Parameters:
  //  - A
  //  - B
  Concat(ctx context.Context, a string, b string) (r *ConcatReply, err error)
}

type AddServiceClient struct {
  c thrift.TClient
}

// Deprecated: Use NewAddService instead
func NewAddServiceClientFactory(t thrift.TTransport, f thrift.TProtocolFactory) *AddServiceClient {
  return &AddServiceClient{
    c: thrift.NewTStandardClient(f.GetProtocol(t), f.GetProtocol(t)),
  }
}

// Deprecated: Use NewAddService instead
func NewAddServiceClientProtocol(t thrift.TTransport, iprot thrift.TProtocol, oprot thrift.TProtocol) *AddServiceClient {
  return &AddServiceClient{
    c: thrift.NewTStandardClient(iprot, oprot),
  }
}

func NewAddServiceClient(c thrift.TClient) *AddServiceClient {
  return &AddServiceClient{
    c: c,
  }
}

// Parameters:
//  - A
//  - B
func (p *AddServiceClient) Sum(ctx context.Context, a int64, b int64) (r *SumReply, err error) {
  var _args0 AddServiceSumArgs
  _args0.A = a
  _args0.B = b
  var _result1 AddServiceSumResult
  if err = p.c.Call(ctx, "Sum", &_args0, &_result1); err != nil {
    return
  }
  return _result1.GetSuccess(), nil
}

// Parameters:
//  - A
//  - B
func (p *AddServiceClient) Concat(ctx context.Context, a string, b string) (r *ConcatReply, err error) {
  var _args2 AddServiceConcatArgs
  _args2.A = a
  _args2.B = b
  var _result3 AddServiceConcatResult
  if err = p.c.Call(ctx, "Concat", &_args2, &_result3); err != nil {
    return
  }
  return _result3.GetSuccess(), nil
}

type AddServiceProcessor struct {
  processorMap map[string]thrift.TProcessorFunction
  handler AddService
}

func (p *AddServiceProcessor) AddToProcessorMap(key string, processor thrift.TProcessorFunction) {
  p.processorMap[key] = processor
}

func (p *AddServiceProcessor) GetProcessorFunction(key string) (processor thrift.TProcessorFunction, ok bool) {
  processor, ok = p.processorMap[key]
  return processor, ok
}

func (p *AddServiceProcessor) ProcessorMap() map[string]thrift.TProcessorFunction {
  return p.processorMap
}

func NewAddServiceProcessor(handler AddService) *AddServiceProcessor {

  self4 := &AddServiceProcessor{handler:handler, processorMap:make(map[string]thrift.TProcessorFunction)}
  self4.processorMap["Sum"] = &addServiceProcessorSum{handler:handler}
  self4.processorMap["Concat"] = &addServiceProcessorConcat{handler:handler}
return self4
}

func (p *AddServiceProcessor) Process(ctx context.Context, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
  name, _, seqId, err := iprot.ReadMessageBegin()
  if err != nil { return false, err }
  if processor, ok := p.GetProcessorFunction(name); ok {
    return processor.Process(ctx, seqId, iprot, oprot)
  }
  iprot.Skip(thrift.STRUCT)
  iprot.ReadMessageEnd()
  x5 := thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function " + name)
  oprot.WriteMessageBegin(name, thrift.EXCEPTION, seqId)
  x5.Write(oprot)
  oprot.WriteMessageEnd()
  oprot.Flush(ctx)
  return false, x5

}

type addServiceProcessorSum struct {
  handler AddService
}

func (p *addServiceProcessorSum) Process(ctx context.Context, seqId int32, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
  args := AddServiceSumArgs{}
  if err = args.Read(iprot); err != nil {
    iprot.ReadMessageEnd()
    x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err.Error())
    oprot.WriteMessageBegin("Sum", thrift.EXCEPTION, seqId)
    x.Write(oprot)
    oprot.WriteMessageEnd()
    oprot.Flush(ctx)
    return false, err
  }

  iprot.ReadMessageEnd()
  result := AddServiceSumResult{}
var retval *SumReply
  var err2 error
  if retval, err2 = p.handler.Sum(ctx, args.A, args.B); err2 != nil {
    x := thrift.NewTApplicationException(thrift.INTERNAL_ERROR, "Internal error processing Sum: " + err2.Error())
    oprot.WriteMessageBegin("Sum", thrift.EXCEPTION, seqId)
    x.Write(oprot)
    oprot.WriteMessageEnd()
    oprot.Flush(ctx)
    return true, err2
  } else {
    result.Success = retval
}
  if err2 = oprot.WriteMessageBegin("Sum", thrift.REPLY, seqId); err2 != nil {
    err = err2
  }
  if err2 = result.Write(oprot); err == nil && err2 != nil {
    err = err2
  }
  if err2 = oprot.WriteMessageEnd(); err == nil && err2 != nil {
    err = err2
  }
  if err2 = oprot.Flush(ctx); err == nil && err2 != nil {
    err = err2
  }
  if err != nil {
    return
  }
  return true, err
}

type addServiceProcessorConcat struct {
  handler AddService
}

func (p *addServiceProcessorConcat) Process(ctx context.Context, seqId int32, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
  args := AddServiceConcatArgs{}
  if err = args.Read(iprot); err != nil {
    iprot.ReadMessageEnd()
    x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err.Error())
    oprot.WriteMessageBegin("Concat", thrift.EXCEPTION, seqId)
    x.Write(oprot)
    oprot.WriteMessageEnd()
    oprot.Flush(ctx)
    return false, err
  }

  iprot.ReadMessageEnd()
  result := AddServiceConcatResult{}
var retval *ConcatReply
  var err2 error
  if retval, err2 = p.handler.Concat(ctx, args.A, args.B); err2 != nil {
    x := thrift.NewTApplicationException(thrift.INTERNAL_ERROR, "Internal error processing Concat: " + err2.Error())
    oprot.WriteMessageBegin("Concat", thrift.EXCEPTION, seqId)
    x.Write(oprot)
    oprot.WriteMessageEnd()
    oprot.Flush(ctx)
    return true, err2
  } else {
    result.Success = retval
}
  if err2 = oprot.WriteMessageBegin("Concat", thrift.REPLY, seqId); err2 != nil {
    err = err2
  }
  if err2 = result.Write(oprot); err == nil && err2 != nil {
    err = err2
  }
  if err2 = oprot.WriteMessageEnd(); err == nil && err2 != nil {
    err = err2
  }
  if err2 = oprot.Flush(ctx); err == nil && err2 != nil {
    err = err2
  }
  if err != nil {
    return
  }
  return true, err
}


// HELPER FUNCTIONS AND STRUCTURES

// Attributes:
//  - A
//  - B
type AddServiceSumArgs struct {
  A int64 `thrift:"a,1" db:"a" json:"a"`
  B int64 `thrift:"b,2" db:"b" json:"b"`
}

func NewAddServiceSumArgs() *AddServiceSumArgs {
  return &AddServiceSumArgs{}
}


func (p *AddServiceSumArgs) GetA() int64 {
  return p.A
}

func (p *AddServiceSumArgs) GetB() int64 {
  return p.B
}
func (p *AddServiceSumArgs) Read(iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 1:
      if fieldTypeId == thrift.I64 {
        if err := p.ReadField1(iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(fieldTypeId); err != nil {
          return err
        }
      }
    case 2:
      if fieldTypeId == thrift.I64 {
        if err := p.ReadField2(iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *AddServiceSumArgs)  ReadField1(iprot thrift.TProtocol) error {
  if v, err := iprot.ReadI64(); err != nil {
  return thrift.PrependError("error reading field 1: ", err)
} else {
  p.A = v
}
  return nil
}

func (p *AddServiceSumArgs)  ReadField2(iprot thrift.TProtocol) error {
  if v, err := iprot.ReadI64(); err != nil {
  return thrift.PrependError("error reading field 2: ", err)
} else {
  p.B = v
}
  return nil
}

func (p *AddServiceSumArgs) Write(oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin("Sum_args"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField1(oprot); err != nil { return err }
    if err := p.writeField2(oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *AddServiceSumArgs) writeField1(oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin("a", thrift.I64, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:a: ", p), err) }
  if err := oprot.WriteI64(int64(p.A)); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.a (1) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:a: ", p), err) }
  return err
}

func (p *AddServiceSumArgs) writeField2(oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin("b", thrift.I64, 2); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:b: ", p), err) }
  if err := oprot.WriteI64(int64(p.B)); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.b (2) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 2:b: ", p), err) }
  return err
}

func (p *AddServiceSumArgs) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("AddServiceSumArgs(%+v)", *p)
}

// Attributes:
//  - Success
type AddServiceSumResult struct {
  Success *SumReply `thrift:"success,0" db:"success" json:"success,omitempty"`
}

func NewAddServiceSumResult() *AddServiceSumResult {
  return &AddServiceSumResult{}
}

var AddServiceSumResult_Success_DEFAULT *SumReply
func (p *AddServiceSumResult) GetSuccess() *SumReply {
  if !p.IsSetSuccess() {
    return AddServiceSumResult_Success_DEFAULT
  }
return p.Success
}
func (p *AddServiceSumResult) IsSetSuccess() bool {
  return p.Success != nil
}

func (p *AddServiceSumResult) Read(iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 0:
      if fieldTypeId == thrift.STRUCT {
        if err := p.ReadField0(iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *AddServiceSumResult)  ReadField0(iprot thrift.TProtocol) error {
  p.Success = &SumReply{}
  if err := p.Success.Read(iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Success), err)
  }
  return nil
}

func (p *AddServiceSumResult) Write(oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin("Sum_result"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField0(oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *AddServiceSumResult) writeField0(oprot thrift.TProtocol) (err error) {
  if p.IsSetSuccess() {
    if err := oprot.WriteFieldBegin("success", thrift.STRUCT, 0); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field begin error 0:success: ", p), err) }
    if err := p.Success.Write(oprot); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Success), err)
    }
    if err := oprot.WriteFieldEnd(); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field end error 0:success: ", p), err) }
  }
  return err
}

func (p *AddServiceSumResult) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("AddServiceSumResult(%+v)", *p)
}

// Attributes:
//  - A
//  - B
type AddServiceConcatArgs struct {
  A string `thrift:"a,1" db:"a" json:"a"`
  B string `thrift:"b,2" db:"b" json:"b"`
}

func NewAddServiceConcatArgs() *AddServiceConcatArgs {
  return &AddServiceConcatArgs{}
}


func (p *AddServiceConcatArgs) GetA() string {
  return p.A
}

func (p *AddServiceConcatArgs) GetB() string {
  return p.B
}
func (p *AddServiceConcatArgs) Read(iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 1:
      if fieldTypeId == thrift.STRING {
        if err := p.ReadField1(iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(fieldTypeId); err != nil {
          return err
        }
      }
    case 2:
      if fieldTypeId == thrift.STRING {
        if err := p.ReadField2(iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *AddServiceConcatArgs)  ReadField1(iprot thrift.TProtocol) error {
  if v, err := iprot.ReadString(); err != nil {
  return thrift.PrependError("error reading field 1: ", err)
} else {
  p.A = v
}
  return nil
}

func (p *AddServiceConcatArgs)  ReadField2(iprot thrift.TProtocol) error {
  if v, err := iprot.ReadString(); err != nil {
  return thrift.PrependError("error reading field 2: ", err)
} else {
  p.B = v
}
  return nil
}

func (p *AddServiceConcatArgs) Write(oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin("Concat_args"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField1(oprot); err != nil { return err }
    if err := p.writeField2(oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *AddServiceConcatArgs) writeField1(oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin("a", thrift.STRING, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:a: ", p), err) }
  if err := oprot.WriteString(string(p.A)); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.a (1) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:a: ", p), err) }
  return err
}

func (p *AddServiceConcatArgs) writeField2(oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin("b", thrift.STRING, 2); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:b: ", p), err) }
  if err := oprot.WriteString(string(p.B)); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.b (2) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 2:b: ", p), err) }
  return err
}

func (p *AddServiceConcatArgs) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("AddServiceConcatArgs(%+v)", *p)
}

// Attributes:
//  - Success
type AddServiceConcatResult struct {
  Success *ConcatReply `thrift:"success,0" db:"success" json:"success,omitempty"`
}

func NewAddServiceConcatResult() *AddServiceConcatResult {
  return &AddServiceConcatResult{}
}

var AddServiceConcatResult_Success_DEFAULT *ConcatReply
func (p *AddServiceConcatResult) GetSuccess() *ConcatReply {
  if !p.IsSetSuccess() {
    return AddServiceConcatResult_Success_DEFAULT
  }
return p.Success
}
func (p *AddServiceConcatResult) IsSetSuccess() bool {
  return p.Success != nil
}

func (p *AddServiceConcatResult) Read(iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin()
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 0:
      if fieldTypeId == thrift.STRUCT {
        if err := p.ReadField0(iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *AddServiceConcatResult)  ReadField0(iprot thrift.TProtocol) error {
  p.Success = &ConcatReply{}
  if err := p.Success.Read(iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Success), err)
  }
  return nil
}

func (p *AddServiceConcatResult) Write(oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin("Concat_result"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField0(oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *AddServiceConcatResult) writeField0(oprot thrift.TProtocol) (err error) {
  if p.IsSetSuccess() {
    if err := oprot.WriteFieldBegin("success", thrift.STRUCT, 0); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field begin error 0:success: ", p), err) }
    if err := p.Success.Write(oprot); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Success), err)
    }
    if err := oprot.WriteFieldEnd(); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field end error 0:success: ", p), err) }
  }
  return err
}

func (p *AddServiceConcatResult) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("AddServiceConcatResult(%+v)", *p)
}


