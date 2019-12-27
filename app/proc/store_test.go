package proc

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/umputun/feed-master/app/feed"

	"github.com/stretchr/testify/assert"
)

const pubDate = "Mon, 02 Jan 2006 15:04:05 -0700"

func TestNewBoltDB(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfile.Name())

	boltDB, err := NewBoltDB(tmpfile.Name())

	assert.Nil(t, err)
	assert.Equal(t, boltDB.DB.Path(), tmpfile.Name())
}

func TestNewBoltDBFileNotExists(t *testing.T) {
	boltDB, err := NewBoltDB("")

	assert.Equal(t, err.Error(), "open : no such file or directory")
	assert.Nil(t, boltDB)
}

func TestSaveIfInvalidPubDate(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfile.Name())

	boltDB, _ := NewBoltDB(tmpfile.Name())

	item := feed.Item{
		PubDate: "100500",
	}

	created, err := boltDB.Save("radio-t", item)

	assert.False(t, created)
	assert.Equal(t, err.Error(), "parsing time \"100500\" as \"Mon, 02 Jan 2006 15:04:05 -0700\": cannot parse \"100500\" as \"Mon\"")
}

func TestSave(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfile.Name())

	boltDB, _ := NewBoltDB(tmpfile.Name())

	item := feed.Item{
		PubDate: pubDate,
	}

	created, err := boltDB.Save("radio-t", item)

	assert.True(t, created)
	assert.Nil(t, err)
}

func TestSaveIfItemIsExists(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfile.Name())

	boltDB, _ := NewBoltDB(tmpfile.Name())

	item := feed.Item{
		PubDate: pubDate,
	}
	created, err := boltDB.Save("radio-t", item)

	created, err = boltDB.Save("radio-t", item)

	assert.False(t, created)
	assert.Nil(t, err)
}

func TestLoadIfNotBucket(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfile.Name())

	boltDB, err := NewBoltDB(tmpfile.Name())

	feedItems, err := boltDB.Load("100500", 5)

	assert.Equal(t, len(feedItems), 0)
	assert.Equal(t, err.Error(), "no bucket for 100500")
}

func TestLoad(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfile.Name())

	boltDB, _ := NewBoltDB(tmpfile.Name())
	boltDB.Save("radio-t", feed.Item{PubDate: pubDate})

	items, err := boltDB.Load("radio-t", 5)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(items))
	assert.Equal(t, items[0].PubDate, pubDate)
}

func TestLoadChackMax(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfile.Name())

	boltDB, _ := NewBoltDB(tmpfile.Name())
	boltDB.Save("radio-t", feed.Item{PubDate: pubDate, GUID: "1"})
	boltDB.Save("radio-t", feed.Item{PubDate: pubDate, GUID: "2"})

	cases := []struct {
		max   int
		count int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{5, 2},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			items, err := boltDB.Load("radio-t", tc.max)

			assert.Nil(t, err)
			assert.Equal(t, tc.count, len(items))
		})
	}
}

func TestBuckets(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfile.Name())
	boltDB, _ := NewBoltDB(tmpfile.Name())

	got, err := boltDB.Buckets()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(got))

	boltDB.Save("radio-t", feed.Item{PubDate: pubDate})
	got, err = boltDB.Buckets()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(got))
}

func TestRemoveOldIfNotExistsBucket(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfile.Name())
	boltDB, _ := NewBoltDB(tmpfile.Name())

	count, err := boltDB.removeOld("radio-t", 5)

	assert.Equal(t, err.Error(), "no bucket for radio-t")
	assert.Equal(t, 0, count)
}

func TestRemoveOld(t *testing.T) {
	cases := []struct {
		keep        int
		countDelete int
	}{
		{0, 2},
		{1, 1},
		{2, 0},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			tmpfile, _ := ioutil.TempFile("", "")
			defer os.Remove(tmpfile.Name())
			boltDB, _ := NewBoltDB(tmpfile.Name())
			boltDB.Save("radio-t", feed.Item{PubDate: pubDate, GUID: "1"})
			boltDB.Save("radio-t", feed.Item{PubDate: pubDate, GUID: "2"})

			count, err := boltDB.removeOld("radio-t", tc.keep)

			assert.Nil(t, err)
			assert.Equal(t, tc.countDelete, count)
		})
	}
}
