Real-time prediction example
============================

Requirements
------------

### Install Vowpal Wabbit

[Installation instruction](https://github.com/JohnLangford/vowpal_wabbit#prerequisite-software)

### Create test model

1. Create dataset:

```
$ touch test.dataset
$ echo "0 example0| 100:1 200:1 300:1 400:1 500:1" >> test.dataset
$ echo "1 example1| 100:1 200:1 350:1 550:1" >> test.dataset
```

2. Training model

```
$ vw -c --passes 20 --holdout_off test.dataset -f test.model -b 28
```

Example of request
------------------

```
$ curl --data "0| 100:1 200:1 300:1 400:1" http://localhost:8080/
```

