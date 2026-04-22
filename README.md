# 6CCS3PRJ
Optimising for LLVM compiler passes using GA approaches.

## Running on Windows
 - Install Docker

```bash
docker build -t <container_name> .
```

```bash
docker run -it --rm -v ${pwd}:/project -w /project <container_name>
```

 - IN the container: go run . 
 - From here you can test running on example projects
 - e.g., (where n is your available time)

```bash
go run . -c "clang++ -I./example/string_concat/include ./example/string_concat/src/TestStringCons.cpp" -t <n>h
```

 - Alternatively, build the project via:

```bash
go build -o <output_filename> .
```

 - And then run via:

```bash
./<output_filename> -c "clang++ -I./example/string_concat/include ./example/string_concat/src/TestStringCons.cpp" -t <n>h
```

 - To review the command line arguments to the executable, use the -help argument:

## Running on MacOS/linux
 - Same as Windows except:
 - {pwd} -> (pwd)
 - e.g. 

```bash
docker build -t foo . && docker run -it --rm -v $(pwd):/project -w /project foo
```

## Execution

Once the run command is executed, the population will be initialised. The genetic algorithm then begins compiling and evaluating the fitness of all the population members. The typical genetic algorithm will then run, choosing the best of the members, performing crossover, and performing mutation. This will repeat until the GA converges or the timeout is reached. Each set of output will be made up of a number of lines, where each line represents a member of the population's passes encoded as integers and fitness score.