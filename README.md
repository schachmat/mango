# Mango

Mango listens on a user-defined Port and serves html formatted manual pages via
http. This can be used only locally (default) with a properly configured
terminal browser as a shell alias to `man` or even in a network to serve all
users in a company. The implementation in Go is pretty simple and minimalistic.

## Dependencies

Mango uses `man` to find the path to a requested man page, `bzcat` to extract it
(let me know or supply a pull request, if your distro uses another compression
algo) and `man2html` to reformat the manpage to a hyperlinked html, which is
then served via http similar to CGI (but it's not CGI).
