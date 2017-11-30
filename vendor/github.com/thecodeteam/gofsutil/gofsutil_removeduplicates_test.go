package gofsutil_test

import (
	"strings"
	"testing"

	"github.com/thecodeteam/gofsutil"
)

func TestRemoveDuplicatesExponentialOrdered_SmallData(t *testing.T) {
	testRemoveDuplicates(
		t,
		gofsutil.RemoveDuplicatesExponentialOrdered,
		benchRemoveDupesData,
		gofsutil.RemoveDuplicates(benchRemoveDupesData))
}

func TestRemoveDuplicatesExponentialUnordered_SmallData(t *testing.T) {
	testRemoveDuplicates(
		t,
		gofsutil.RemoveDuplicatesExponentialUnordered,
		benchRemoveDupesData,
		gofsutil.RemoveDuplicates(benchRemoveDupesData))
}

func TestRemoveDuplicatesLinearOrdered_SmallData(t *testing.T) {
	testRemoveDuplicates(
		t,
		gofsutil.RemoveDuplicatesLinearOrdered,
		benchRemoveDupesData,
		gofsutil.RemoveDuplicates(benchRemoveDupesData))
}

func TestRemoveDuplicatesExponentialOrdered_BigData(t *testing.T) {
	testRemoveDuplicates(
		t,
		gofsutil.RemoveDuplicatesExponentialOrdered,
		strings.Fields(loremIpsum),
		gofsutil.RemoveDuplicates(strings.Fields(loremIpsum)))
}

func TestRemoveDuplicatesExponentialUnordered_BigData(t *testing.T) {
	testRemoveDuplicates(
		t,
		gofsutil.RemoveDuplicatesExponentialUnordered,
		strings.Fields(loremIpsum),
		gofsutil.RemoveDuplicates(strings.Fields(loremIpsum)))
}

func TestRemoveDuplicatesLinearOrdered_BigData(t *testing.T) {
	testRemoveDuplicates(
		t,
		gofsutil.RemoveDuplicatesLinearOrdered,
		strings.Fields(loremIpsum),
		gofsutil.RemoveDuplicates(strings.Fields(loremIpsum)))
}

func BenchmarkRemoveDuplicates_Exponential_Ordered___SmallData(b *testing.B) {
	benchmarkRemoveDuplicates(
		b,
		gofsutil.RemoveDuplicatesExponentialOrdered,
		benchRemoveDupesData)
}

func BenchmarkRemoveDuplicates_Exponential_Unordered_SmallData(b *testing.B) {
	benchmarkRemoveDuplicates(
		b,
		gofsutil.RemoveDuplicatesExponentialUnordered,
		benchRemoveDupesData)
}

func BenchmarkRemoveDuplicates_Linear______Ordered___SmallData(b *testing.B) {
	benchmarkRemoveDuplicates(
		b,
		gofsutil.RemoveDuplicatesLinearOrdered,
		benchRemoveDupesData)
}

func BenchmarkRemoveDuplicates_Exponential_Ordered___BigData(b *testing.B) {
	benchmarkRemoveDuplicates(
		b,
		gofsutil.RemoveDuplicatesExponentialOrdered,
		strings.Fields(loremIpsum))
}

func BenchmarkRemoveDuplicates_Exponential_Unordered_BigData(b *testing.B) {
	benchmarkRemoveDuplicates(
		b,
		gofsutil.RemoveDuplicatesExponentialUnordered,
		strings.Fields(loremIpsum))
}

func BenchmarkRemoveDuplicates_Linear______Ordered___BigData(b *testing.B) {
	benchmarkRemoveDuplicates(
		b,
		gofsutil.RemoveDuplicatesLinearOrdered,
		strings.Fields(loremIpsum))
}

func testRemoveDuplicates(t *testing.T, f rdf, d, e []string) {
	// Execute the function that removes the duplicates.
	r := f(d)

	// Validate the result.
	le := len(e)
	tn := t.Name()
	if l := len(r); l != len(e) {
		t.Logf("%s: invalid len: exp=%d, act=%d: %v", tn, le, l, r)
		t.Fail()
	}

	if strings.Contains(tn, "Unordered") {
		return
	}

	for i, v := range r {
		if v != e[i] {
			t.Logf("%s: invalid data: i=%d, exp=%s, act=%s", tn, i, e[i], v)
			t.Fail()
		}
	}
}

func benchmarkRemoveDuplicates(b *testing.B, f rdf, d []string) {
	// Create copies of the data to benchmark.
	c := make([][]string, b.N)
	for i := 0; i < len(c); i++ {
		c[i] = make([]string, len(d))
		copy(c[i], d)
	}

	// Perform the benchmark.
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Execute the function that removes the duplicates.
		f(c[i])
	}
}

var benchRemoveDupesData = []string{
	"a", "b", "b", "c", "cat", "", "cat2", "dog", "dog", "z", ""}

type rdf func([]string) []string

const loremIpsum = `Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Nullam vestibulum porttitor placerat. Praesent ac lorem mauris. Pellentesque
ultrices, nibh dapibus luctus vulputate, erat ligula lacinia massa, nec
pulvinar elit urna in massa. Interdum et malesuada fames ac ante ipsum primis
in faucibus. Vestibulum ut felis vel turpis interdum interdum. Aenean semper
tempus mattis. Nulla justo libero, pharetra et nulla in, euismod faucibus
dolor. Quisque fringilla, nibh in sollicitudin facilisis, lectus odio semper
nisi, non aliquet nunc arcu non nulla. Etiam facilisis libero vel libero
viverra, vel vulputate neque congue. Ut fermentum quam eget nunc sollicitudin
auctor.

Nam rhoncus imperdiet interdum. Vestibulum facilisis odio dictum velit
condimentum, eu congue nibh gravida. Nulla rutrum eros porttitor eros
suscipit, a vulputate nulla semper. Praesent molestie sollicitudin tincidunt.
Phasellus venenatis mattis mauris, sed fermentum libero ornare quis. Curabitur
mauris odio, posuere ac tincidunt at, posuere a tortor. Pellentesque volutpat
erat ac maximus maximus. Mauris dolor sapien, aliquet non hendrerit eu,
ultrices non arcu. Quisque faucibus eros viverra lacus consectetur iaculis
quis ac ex. In vitae pretium risus, quis fermentum lectus. Etiam vestibulum,
odio in laoreet dapibus, ex nisi pellentesque arcu, vel mattis tortor turpis
nec arcu. Cras scelerisque quam vitae convallis vestibulum. Donec suscipit
urna odio, at luctus libero cursus interdum. Cras mollis bibendum auctor.

Cras tempor, lacus nec sodales facilisis, nisl enim consequat massa, at
efficitur lectus metus sed augue. Nulla vitae risus lacinia, mattis ex vel,
vulputate dui. Nunc id sollicitudin nisl. Aenean a sapien elit. Morbi eget
neque vulputate, consectetur augue eget, consectetur turpis. Curabitur porta,
purus a ultrices bibendum, mauris ipsum posuere dui, at gravida tortor augue
in velit. In mollis enim eu malesuada tristique. Ut dui lectus, aliquet et
felis vel, tristique ultrices nunc. Nunc luctus iaculis ipsum at luctus. Donec
maximus tellus id justo cursus, eu rhoncus diam aliquet. Nam sit amet hendrerit
massa. Fusce enim libero, tincidunt eget nisi at, viverra lobortis enim.

Ut eleifend, nunc in molestie mollis, erat urna interdum neque, id sagittis
ex nisi at libero. Vivamus auctor leo augue, eget ultricies elit dictum eget.
Nam lacinia, urna vel vestibulum placerat, nisl eros accumsan eros, et dictum
nisl nibh ut nisl. Maecenas nec lobortis libero. Vestibulum non vestibulum
quam. Nulla feugiat dignissim augue, eget faucibus metus porta a. Pellentesque
habitant morbi tristique senectus et netus et malesuada fames ac turpis
egestas. Pellentesque lobortis suscipit aliquet. Aenean mattis est velit,
sed consequat lectus ultrices id. Phasellus vel massa ut quam accumsan
pellentesque. Nullam id porttitor orci. Suspendisse suscipit hendrerit tempor.

Mauris finibus lorem nibh, et tristique nunc consequat et. Maecenas sodales
dolor vitae dui fringilla auctor. Pellentesque vitae tempus diam, sed bibendum
risus. Sed volutpat, ligula id iaculis hendrerit, ex eros dapibus risus, ut
scelerisque tellus ipsum ac tellus. Nulla facilisi. Suspendisse ut
pellentesque ex. Phasellus molestie est ac accumsan egestas. Nullam dignissim,
sapien id tempor tincidunt, leo odio dictum ligula, vel lacinia metus turpis
sit amet est. Curabitur mattis dignissim ipsum nec ullamcorper. Proin sagittis
sem arcu, a auctor neque ultrices tincidunt. Etiam luctus elementum pulvinar.
Donec ornare, mauris eget porttitor maximus, dolor turpis rhoncus felis, quis
pretium mi nibh id justo. Proin id pharetra est. Donec auctor tortor eu metus
vehicula, ac hendrerit dui porta. Sed tellus purus, vestibulum a neque in,
eleifend tempor arcu.`
