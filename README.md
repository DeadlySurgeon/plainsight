# Plain Sight

Something terrible is in (potentially) plain sight. This PoC proves that not all
code is as simple as it might seem and expands upon an example snippet I
produced that ended up plastering a "You lost the game!" banner followed by a
"you looked" gesturing ASCII art hand to the terminal.

While the previous iteration was more childish and mundane, this example builds
upon that to showcase a concern: It is relatively easy to hide such a nasty
bit of binary, and an attacker with significantly more patience than I have for
this example could craft something even more horrifying. This attack vector is
also not inherently specific to Go. But something also to point out, even with a
100% coverage rate for the entire code base, this code actually redirects
whatever username and password to an adversary.

```sh
$ go run main.go --username=bob --password=pass
```

Although the default URL is defined as `http://provider.cluster.local`, running
the command above shows that the application actually connects to
`https://evildomain.lab` which you won't find anywhere in this repository,
besides for just now in this very README.md file.

## Coverage Percentage

Not to rant, but while I am here, this repository has **a 100% coverage rate**.
This is the exact reason why I think unit tests for the sake of percentage is
an awful idea. It is so easy to overlook something critically important, and
even more so you can have a 100% coverage rate while missing significant amounts
of input. Also, since coverage targeting percentage tends to be more about line
coverage than behavior coverage, when it is time to change those lines or add to
them you'll find that you'll have to touch the test or maybe even refactor it
entirely. Thus, by writing unit tests for the sake of Coverage Percentage
instead of to a Defined Behavior (fuzzy or not), you end up trading flexibility
with a false metric.

If you or a loved one is pushing this on your project, please seek out
professional help. It is never too late.

```sh
$ go test ./... --coverprofile coverage.out
ok      github.com/deadlysurgeon/plainsight             0.007s  coverage: 100.0% of statements
ok      github.com/deadlysurgeon/plainsight/auth        0.024s  coverage: 100.0% of statements
```

## License

```
Copyright 2024 deadly.surgery

Licensed under the Apache License, Version 2.0 (the "License");
you may not use these files except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
