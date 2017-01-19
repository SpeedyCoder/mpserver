# mpserver

Assumptions:
* Component terminates normally only if its input channel is closed.
* Before terminating, every Component closes its output channel.