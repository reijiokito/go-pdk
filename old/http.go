package old

//
//import (
//	"github.com/reijiokito/go-pdk"
//	"sync"
//)
//
//type HttpContext struct {
//	Request  go_pdk.Request
//	Response go_pdk.Response
//}
//
//type HttpContextPool struct {
//	sync.Pool
//}
//
//func (ctx *HttpContext) reset() {
//	ctx.Request.Body = ctx.Request.Body[0:0]
//	ctx.Response.Body = ctx.Response.Body[0:0]
//	ctx.Response.Code = int32(0)
//}
//
//func newHttpContext() *HttpContext {
//	return &HttpContext{
//		Request: go_pdk.Request{
//			Body: make([]byte, 16384),
//		},
//		Response: go_pdk.Response{
//			Body: make([]byte, 0),
//			Code: 200,
//		},
//	}
//}
//
//func (p *HttpContextPool) GetContext() *HttpContext {
//	service, _ := p.Get().(*HttpContext)
//	return service
//}
//
//func (p *HttpContextPool) PutContext(service *HttpContext) {
//	service.reset()
//	p.Put(service)
//}
//
//func NewContextPool() HttpContextPool {
//	return HttpContextPool{
//		sync.Pool{
//			New: func() interface{} { return newHttpContext() },
//		}}
//}
//
////
////func (r *go_pdk.Request) Build(req *http.Request) error {
////	if req == nil {
////		return errors.New("natsproxy: Request cannot be nil")
////	}
////
////	buf := bytes.NewBuffer(r.Body)
////	buf.Reset()
////	if req.Body != nil {
////		if _, err := io.Copy(buf, req.Body); err != nil {
////			return err
////		}
////		if err := req.Body.Close(); err != nil {
////			return err
////		}
////	}
////
////	r.Body = buf.Bytes()
////	return nil
////}
