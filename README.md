# r2n

[task](https://github.com/go-task/task) 에서 `output: prefixed` 사용시
`curl -# ...` 과 같이 stderr에 \n 대신 \r를 내보내서 라인 단위 집계가 되지 않는 경우
task는 curl이 종료되어서 \n이 올때까지 stderr를 버퍼링하고 출력하지 않는다.
그럼 curl의 progress bar가 전혀 보이지 않기 때문에, r2n으로 \r > \n 강제 변환을 해준다.
r2n은 stderr의 \r만 \n로 변환한다. (stdout은 pipe 사용시 데이터가 깨질 위험이 있음)

```sh
go install https://github.com/dgdsingen/go/r2n@latest
r2n curl -L "https://test.com/test.tar.gz" | tar --totals -xzf -
```
