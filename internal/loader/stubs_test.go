package loader

var (
	pageOK = []byte(`<!DOCTYPE html>

<html lang="en-US">

<head>
    <meta charset="utf-8">
	<base href="http://test.com"/>
</head>

<body>
	<a href="http://abs.link.com">Abs link</a>
	<a href="rel/link">Abs link</a>
<body/>

<html/>

`)

	nonHTMLPage = []byte(`Some random non-HTML payload`)

	badHTMLPage = []byte(`<!DOCTYPE html>

<html lang="en-US">

<head>
    <meta charset="utf-8">
	<base href="http://test.com"/>
	<base href="http://another.test.com"/>
</head>

<body>
	<a href="http://abs.link.com">Abs link</a>
	<a href="rel/link">Abs link</a>
<body/>

<html/>
`)

	pageWithTwoBases = []byte(`<!DOCTYPE html>

<html lang="en-US">

<head>
    <meta charset="utf-8">
	<base href="http://test.com"/>
	<base href="http://another.com"/>
</head>

<body>
	<a href="http://abs.link.com">Abs link</a>
	<a href="rel/link">Abs link</a>
<body/>

<html/>

`)
)
