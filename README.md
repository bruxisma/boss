Overview
========

> *There is no such thing as luck on shared hosting*
>   -- The Boss, MGS3 (or, like, whatever)

Boss is a small tool meant to be used as a git pre-receive hook.  Specifically,
it was created for use with Dreamhost's shared hosting and [hugo][1] so that a
`git push` becomes your deploy command.

Installation
------------

Please ensure that [hugo][2] is installed and available on your system PATH
before installing Boss. Future versions will permit specifying its location via
the configuration file, however this initial release expects it to be on the
system PATH.If you have not yet setup your bare git repo, perform the following
(replace `<location>` with your own name):

```sh
git init --bare <location>.git
go get -u -v -ldflags "-o <location>.git/hooks/pre-receive" github.com/slurps-mad-rips/boss 
```

If you receive an error regarding "`a.out` no such file or directory", feel
free to ignore it. The executable was placed into the correct location.

Please make sure to add a configuration file with the destination of your site.

Configuration
-------------

Place any [viper][2] parsable config file at `$HOME/.config/boss.<ext>`. The
configuration has several values for tweaking, however they are rarely needed.
The exception here is the `destination` key. If on Dreamhost, it should be set
to `$HOME/<website>.<tld>`. Boss will handle the rest. Then, perform your
typical `git push` workflow to the Dreamhost server, and watch the output from
hugo. The full configuration 'schema' is found below in the yaml format:

```
destination: REQUIRED # This can use environment variables
# optional flags, defaults displayed
build-cache: $HOME/.cache/hugo
branch: master
clean-destination-dir: true
minify: true
gc: true
```

[1]: https://gohugo.io
[2]: https://github.com/spf13/viper
