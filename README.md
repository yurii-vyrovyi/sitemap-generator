# sitemap-generator

`sitemap-generator` is a tool for sitemaps generating (obviously).

### How to run it?

```usage
sitemap-generator <url> [-parallel=...] [-output-file=...] [-max-depth=...]
url				an url of website you want to build sitemap of

optional
	-parallel=		number of parallel workers to navigate through site
	-output-file=		output file path
	-max-depth=		max depth of url navigation recursion

```

### Protocol
A protocol description could be found [here](https://sitemaps.org/protocol.html).

### Notes
It was created as a test task. While it works in general and shows approaches there aree several weak points that need to be improved to make this sitemap generator a real tool:
- escaping is done partially
- resulting file size is not limited
