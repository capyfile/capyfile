# Performance

Capyfile can work in the concurrent mode, which can increase the performance of
the pipeline. 

The concurrent mode is enabled by using the `-c` or `--concurrency` option. 

You also can specify which concurrency mode to use by setting the `-m` or 
`--concurrency-mode` option. Those concurrency modes are available:
* `event` - uses the event-based concurrency algorithm (default)
* `lock` - uses the lock-based concurrency algorithm

## Which concurrency mode to use?

The short answer is: it depends. It depends on the number of files, the size of
the files, the number of operations and their complexity. So, try to experiment
with both modes to find the best one for your pipeline.

Here are some examples of when to use each mode.

**Example 1:** 500 files of 1-10 KB each.

Go benchmark in the **event-based** concurrency mode:
```
$ go test -bench=. -count 10 -benchmem
goos: linux
goarch: amd64
pkg: capyfile/capysvc
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              850           4133325 ns/op         2582354 B/op      36806 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              340           3424421 ns/op         2593162 B/op      36819 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              493           3673923 ns/op         2587335 B/op      36812 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              490           3626531 ns/op         2587581 B/op      36812 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              410           3554254 ns/op         2590542 B/op      36816 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              427           3440455 ns/op         2590134 B/op      36815 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              433           3316896 ns/op         2588851 B/op      36815 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              381           3085912 ns/op         2591981 B/op      36816 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              372           3147571 ns/op         2590995 B/op      36817 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              454           3245242 ns/op         2588073 B/op      36813 allocs/op
```

Go benchmark in the **lock-based** concurrency mode:
```
$ go test -bench=. -count 10 -benchmem
goos: linux
goarch: amd64
pkg: capyfile/capysvc
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               618           1886870 ns/op         2577517 B/op      36778 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               632           1935147 ns/op         2577309 B/op      36778 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               624           1943047 ns/op         2577355 B/op      36777 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               595           1912312 ns/op         2577836 B/op      36778 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               630           1861319 ns/op         2577224 B/op      36777 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               620           1897534 ns/op         2577510 B/op      36778 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               607           1912895 ns/op         2577665 B/op      36778 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               634           1913610 ns/op         2577257 B/op      36778 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               595           1928330 ns/op         2577773 B/op      36777 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               657           1903118 ns/op         2576822 B/op      36776 allocs/op
```

As you can see, the lock-based concurrency mode is faster than the event-based
concurrency mode.

**Example 2:** 50 files of 1-10 KB each.

Go benchmark in the **event-based** concurrency mode:
```
$ go test -bench=. -count 10 -benchmem
goos: linux
goarch: amd64
pkg: capyfile/capysvc
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             3013            797971 ns/op          269334 B/op       3831 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             5488            552328 ns/op          269257 B/op       3831 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             6543            465441 ns/op          269415 B/op       3831 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             5830            507417 ns/op          269164 B/op       3831 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             5752            556283 ns/op          269258 B/op       3831 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             5868            601361 ns/op          269358 B/op       3831 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             5640            616318 ns/op          269555 B/op       3831 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             5299            678332 ns/op          268789 B/op       3831 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             5746            644208 ns/op          269716 B/op       3831 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             5628            699824 ns/op          268783 B/op       3831 allocs/op
```

Go benchmark in the **lock-based** concurrency mode:
```
$ go test -bench=. -count 10 -benchmem
goos: linux
goarch: amd64
pkg: capyfile/capysvc
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkService_RunProcessorConcurrentlyInLockMode-24              1879            594739 ns/op          261257 B/op       3792 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24              3098            566712 ns/op          261117 B/op       3792 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24              2512            492381 ns/op          261149 B/op       3792 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24              2880            505255 ns/op          261123 B/op       3792 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24              2745            642391 ns/op          261133 B/op       3792 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24              2414            616101 ns/op          261168 B/op       3792 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24              2322            664200 ns/op          261183 B/op       3792 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24              2679            610138 ns/op          261145 B/op       3792 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24              2084            645171 ns/op          261201 B/op       3792 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24              2103            614091 ns/op          261212 B/op       3792 allocs/op
```

As you can see, the event-based concurrency mode is much faster than the
lock-based.

**Example 3:** 500 files of 0.1-1 MB each.

Go benchmark in the **event-based** concurrency mode:
```
$ go test -bench=. -count 10 -benchmem
goos: linux
goarch: amd64
pkg: capyfile/capysvc
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              381           2786374 ns/op         2571669 B/op      12197 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              346           3486458 ns/op         2714479 B/op      12201 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              338           3172617 ns/op         2750888 B/op      12199 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              370           3450970 ns/op         2614032 B/op      12198 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              556           3900825 ns/op         2128994 B/op      12190 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              298           4148192 ns/op         2963367 B/op      12203 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              256           4221964 ns/op         3258762 B/op      12208 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              228           4393880 ns/op         3516014 B/op      12212 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              258           4376282 ns/op         3242436 B/op      12207 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24              273           4370870 ns/op         3128253 B/op      12205 allocs/op
```

Go benchmark in the **lock-based** concurrency mode:
```
$ go test -bench=. -count 10 -benchmem
goos: linux
goarch: amd64
pkg: capyfile/capysvc
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               894           1197198 ns/op         1755931 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               963           1175468 ns/op         1712931 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               988           1148555 ns/op         1698843 B/op      12146 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               928           1180475 ns/op         1733944 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               960           1191753 ns/op         1714680 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               967           1191474 ns/op         1710678 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               949           1193231 ns/op         1721166 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               908           1201709 ns/op         1746702 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               940           1172060 ns/op         1726570 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               951           1174498 ns/op         1719979 B/op      12147 allocs/op
```

The lock-based concurrency mode is clear winner in this case.

**Example 4:** 50 files of 0.1-1 MB each.

Go benchmark in the **event-based** concurrency mode:
```
$ go test -bench=. -count 10 -benchmem
goos: linux
goarch: amd64
pkg: capyfile/capysvc
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             3765            399965 ns/op          143794 B/op       1314 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             2528            475484 ns/op          150775 B/op       1315 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             3162            354947 ns/op          146546 B/op       1315 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             4446            322204 ns/op          141641 B/op       1315 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             4376            297873 ns/op          141698 B/op       1315 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             3907            299362 ns/op          143175 B/op       1315 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             3708            313741 ns/op          143907 B/op       1315 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             3212            370670 ns/op          146682 B/op       1315 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             3481            351267 ns/op          144847 B/op       1315 allocs/op
BenchmarkService_RunProcessorConcurrentlyInEventMode-24             3290            305308 ns/op          146415 B/op       1315 allocs/op
```

Go benchmark in the **lock-based** concurrency mode:
```
$ go test -bench=. -count 10 -benchmem
goos: linux
goarch: amd64
pkg: capyfile/capysvc
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               915           1181075 ns/op         1742175 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               972           1133733 ns/op         1707790 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               985           1155578 ns/op         1700525 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               916           1181013 ns/op         1741565 B/op      12148 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               966           1196290 ns/op         1711236 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               939           1162089 ns/op         1727162 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               963           1198022 ns/op         1712967 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               970           1163891 ns/op         1708937 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               982           1156418 ns/op         1702189 B/op      12147 allocs/op
BenchmarkService_RunProcessorConcurrentlyInLockMode-24               962           1171188 ns/op         1713511 B/op      12147 allocs/op
```

In this case, the event-based concurrency mode is much faster than the
lock-based.