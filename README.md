# parallel-shell-actions

```shell
// proposed usage
name: Proposed calling convention

jobs:
  some-job:
    runs-on: ubuntu-latest
    steps:
    - .....
    - .....
    - name: "run (a, b, c) and (d, e, f) concurrently"
      uses: thisrepo/parallel-shell-actions@v1
      with:
        commands: |
          [
            {
              "shell": "bash",
              "commands": [
                "echo a",
                "echo b",
                "echo c"
              ]
            },
            {
              "shell": "bash",
              "commands": [
                "echo d",
                "echo e",
                "echo f"
              ]
            }
          ]
    - .....
```
