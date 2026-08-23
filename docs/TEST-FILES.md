# Test Files

When enabled, the installer asks for sizes in MB as plain numbers, for example:

```text
100 500 1000
```

Files are created under `/var/www/routeglass-testfiles/` and exposed through `/testfiles/`, for example:

```text
https://your-domain.example/testfiles/100MB.bin
```

Sparse file creation is used to avoid unnecessarily writing the entire requested size during installation.
