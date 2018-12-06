# loco

An utility to collect logs.

In the last weeks I've found myself using `logrotate` to organize log files of some scripts and tools... It's easy and flexible, but I find it a bit cumbersome: create a configuration file, create a `cron` entry... so I decided to code my own tool (mainly because I wanted to practice a bit with Go). `loco` capures standard out and redirect it to a file, rotating it at a specified interval.

# Usage

In its simpler form you can use `loco` like this:

```bash
$ some-command | loco collect /path/to/log/file
```

`loco` will rotate the log file every day. If you want to use a different interval you can:

* Create a configuration for a log file. For instance, if you want the log file to be rotated every week:

  ```bash
  $ loco config -i 1w /path/to/log/file.log
  ```

* Change the defaults; if you want to set all the log files rotate, by default, every 3 days:

  ```bash
  $ loco defaults -i 3d
  ```

* Set the `LOCO_INTERVAL` environment variable to a valid interval

## Valid intervals

Valid intervals have the form `\d+[mhdwM]`

* `m` stands for minute
* `h` stands for hour
* `d` stands for days
* `w` stands for weeks
` `M` stands for months

## Configurations

To create or edit configurations:

```bash
$ loco config -i <interval> /path/to/file.log
```

To list active configurations:

```bash
$ loco list
```

To remove a configuration (not the log files):

```bash
$ loco remove /path/to/file.log
```

## Defaults

To show the defaults:

```bash
$ loco defaults
```

To set defaults:

```bash
$ loco defaults -i <interval>
```

# Collecting logs

```bash
$ loco collect /path/to/file.log
```

The `-t` or `--tee` makes `loco` work as the `tee` command: output is send to both log file and stdout.

# TODO

* Custom suffix
* Save rotate history in state?
* Max rotations
