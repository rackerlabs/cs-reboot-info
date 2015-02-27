# cs-reboot-info tool

This is a tool that will allow you to quickly identify servers affected by the upcoming reboot. It is fully OS-independent and easy to use.

## Installation

If you are a Linux or OSX user, you will need to download the binary like so:

```bash
wget https://github.com/blah
```

If you are a Windows user, you will need to enter the following link in your browser:

> https://github.com/blah

and download the file. When the download is finished, navigate to the folder where the download is located.

## Usage

The list ALL the servers which will be affected by the upcoming reboot, you will need to run:

```bash
./cs-reboot-info
```

This is the default and simplest usage. You can also modify behaviour by these options:

### --csv

will output everything into a local CSV file in the same directory called `cs-reboot-info-output.csv`. This flag is optional.

## --localtime

will convert all of the UTC times to your local timezone. 
