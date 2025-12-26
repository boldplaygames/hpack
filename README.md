# hpack

hpack은 msgpack을 변형한 패킷 Serializer입니다. <br>
struct 필드명을 해시코드로 변환하여 직렬화합니다.


## Repository visibility
🚨 Public

## Hashing field name
어플리케이션에서 정의한 문자열 필드명을 CRC32로 해싱 후 XOR 연산으로 크기를 줄여서 각 필드에 할당합니다.
다만 인코딩에 사용하는 해시코드 사이즈는 struct내 필드명 해시코드들의 최대 크기로 결정합니다(즉, struct 내 모든 필드들의 사이즈는 동일합니다).


## What is diffrent from msgpack
### 1. field name type
||Description|
|--|--|
|msgpack|어플리케이션에서 정의한 문자열 필드명|
|hpack |필드명을 CRC32로 해싱 후 XOR 연산으로 사이즈를 줄인 Hash code|


### 2. struct 타입 포맷
#### msgpack
https://github.com/msgpack/msgpack/blob/master/spec.md#map-format-family
``` 
fixmap stores a map whose length is upto 15 elements
+--------+~~~~~~~~~~~~~~~~~+
|1000XXXX|   N*2 objects   |
+--------+~~~~~~~~~~~~~~~~~+

map 16 stores a map whose length is upto (2^16)-1 elements
+--------+--------+--------+~~~~~~~~~~~~~~~~~+
|  0xde  |YYYYYYYY|YYYYYYYY|   N*2 objects   |
+--------+--------+--------+~~~~~~~~~~~~~~~~~+

map 32 stores a map whose length is upto (2^32)-1 elements
+--------+--------+--------+--------+--------+~~~~~~~~~~~~~~~~~+
|  0xdf  |ZZZZZZZZ|ZZZZZZZZ|ZZZZZZZZ|ZZZZZZZZ|   N*2 objects   |
+--------+--------+--------+--------+--------+~~~~~~~~~~~~~~~~~+

where
* XXXX is a 4-bit unsigned integer which represents N
* YYYYYYYY_YYYYYYYY is a 16-bit big-endian unsigned integer which represents N
* ZZZZZZZZ_ZZZZZZZZ_ZZZZZZZZ_ZZZZZZZZ is a 32-bit big-endian unsigned integer which represents N
* N is the size of a map
* odd elements in objects are keys of a map
* the next element of a key is its associated value 
```
<br>

#### hpack
 msgpack과 달리 객체 앞에 필드명 해시코드의 사이즈를 추가

```
# mapLen, N*2 objects: msgpack과 동일

필드명이 1byte인 경우
+========+--------+~~~~~~~~~~~~~~~~~+
| mapLen |  0x00  |   N*2 objects   |
+========+--------+~~~~~~~~~~~~~~~~~+

필드명이 2byte인 경우
+========+--------+~~~~~~~~~~~~~~~~~+
| mapLen |  0x40  |   N*2 objects   |
+========+--------+~~~~~~~~~~~~~~~~~+

필드명이 4byte인 경우
+========+--------+~~~~~~~~~~~~~~~~~+
| mapLen |  0x80  |   N*2 objects   |
+========+--------+~~~~~~~~~~~~~~~~~+
```

## Reference
### msgpack 
[github.com/vmihailenco/msgpack/v5 v5.4.1](https://pkg.go.dev/github.com/vmihailenco/msgpack/v5@v5.4.1)