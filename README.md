# r2n

r2n은 stderr를 읽어 `\r` 를 `\n` 로 변환한다. stdout은 pipe 사용시 데이터가 깨질 위험이 있으므로 변환하지 않는다.

## why?

[task](https://github.com/go-task/task) 에서 `output: interleaved` (default) 설정 + task 병렬 처리시 로그가 뒤섞인다. \
이때 `output: prefixed` 설정해주면 stdout/stderr를 버퍼링해서 라인 단위로 로그 앞에 `[task]` prefix를 붙여줘서 로그 구분이 용이해진다.

다만 라인 단위로 집계가 바뀌기 때문에 `curl -# ...` 과 같이 stderr에 `\n` 대신 `\r` 를 내보내는 경우 라인 단위 집계가 되지 않아 \
curl이 종료되어 `\n` 이 올때까지 stderr가 버퍼링되며 출력되지 않는다. 즉 progress가 전혀 보이지 않게 된다.

이걸 해결하려면 다시 `output: interleaved` 로 돌아가거나 task 별로 `interactive: true` 를 붙여줘야 한다. \
그럼 다시 로그가 뒤섞이는 이슈가 있으므로 두 단점 중 하나를 안고 갈 수 밖에 없다.

r2n은 stderr를 읽어 `\r` 를 `\n` 로 변환해줌으로서 `output: prefixed` 와 실시간 출력을 동시에 가능하게 해준다.

## install

```sh
GOPROXY=direct go install github.com/dgdsingen/go/cmd/r2n@latest
```

## usage

```sh
r2n curl -L "https://test.com/test.tar.gz" | tar --totals -xzf -
```
