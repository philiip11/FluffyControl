# FluffyControl
A simple server for controlling an Samsung Navibot SR8855 (or similar device) via Google Assistant or other REST-Calls.

# Why
The Samsung Navibot SR8855 has a timer function to run every day at a given time.
When the battery runs low, it will return to charge once and then carries on. *But*, what if the battery was low before starting?
Then it cleans for about 5 minutes, charges for about 20 minutes and cleans until the battery dies again or it cleaned everywhere.
If your flat is big enough, that you need two complete runs of cleaning, this is quite impractical.

This tool takes control over your Samsung Navibot and takes care of the battery problem.
Also, it provides a REST-API for connecting Google Assistant via ifttt.com to your vacuum cleaner.

# Installation
You will need a Raspberry Pi with the latest Raspbian image, an IR LED and, if you want to use another vacuum cleaner, an IR receiver to record the IR codes.
If you want to use this server with a Samsung Navibot SR8855, you can copy the `lircd.conf` file to `/etc/lirc/lircd.conf`.

Install the IR LED and receiver according to this instructions:
[Setting up lirc](https://github.com/AnaviTechnology/anavi-docs/blob/master/anavi-infrared-phat/anavi-infrared-phat.md#setting-up-lirc)

// TODO

// Add more instructions
