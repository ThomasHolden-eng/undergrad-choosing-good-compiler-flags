# 6CCS3PRJ
Optimising for LLVM compiler passes using GA approaches.

## Running on Windows
 - Install Docker
 - docker build -t <container_name> .
 - docker run -it --rm -v ${pwd}:/project -w /project <container_name>
 - IN the container: go run . 
 - From here you can test running on example projects

## Running on MacOS/linux
 - Same as Windows except:
 - {pwd} -> (pwd)
 - e.g. 
```
docker build -t foo . && docker run -it --rm -v $(pwd):/project -w /project foo
```