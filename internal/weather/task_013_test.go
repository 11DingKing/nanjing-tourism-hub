package weather
import("context";"io";"net/http";"strings";"testing";"time")
type task013Body struct{io.Reader;closed bool}
func(b *task013Body)Close()error{b.closed=true;return nil}
type task013Transport func(*http.Request)(*http.Response,error)
func(f task013Transport)RoundTrip(r *http.Request)(*http.Response,error){return f(r)}
func TestTask013(t *testing.T){body:=&task013Body{Reader:strings.NewReader("ok")};p:=New("http://weather");p.HTTP=&http.Client{Transport:task013Transport(func(*http.Request)(*http.Response,error){return &http.Response{StatusCode:200,Body:body,Header:make(http.Header)},nil})};if _,e:=p.Forecast(context.Background(),"r1",time.Now());e!=nil{t.Fatal(e)};if !body.closed{t.Fatal("weather response body left open")}}