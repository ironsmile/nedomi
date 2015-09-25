package app

/*
func TestRestoreFromDisk(t *testing.T) {
	t.Parallel()
	expected := "awesome"
	up, cz, _, _ := setup()
	ca := NewFakeCacheAlgorithm()
	runtime.GOMAXPROCS(runtime.NumCPU())
	up.addFakeResponse("/path",
		fakeResponse{
			Status:       "200",
			ResponseTime: 0,
			Response:     expected,
		})

	storage := New(cz, ca, newStdLogger())
	defer os.RemoveAll(storage.path)

	ctx := contexts.NewVhostContext(context.Background(), &types.VirtualHost{Upstream: up})
	oid := types.ObjectID{
		CacheKey: "1",
		Path:     "/path",
	}
	index := types.ObjectIndex{
		ObjID: oid,
		Part:  0,
	}
	ca.AddFakeReplies(index, FakeReplies{
		Lookup:     LookupFalse,
		ShouldKeep: ShouldKeepTrue,
		AddObject:  AddObjectNil,
	})

	makeAndCheckGetFullFile(t, ctx, storage, oid, expected)

	if err := storage.Close(); err != nil {
		t.Errorf("Storage.Close() shouldn't have returned error\n%s", err)
	}

	ca.AddFakeReplies(index, FakeReplies{
		Lookup:     LookupTrue, // change to True now that we supposevely will read it from disk
		ShouldKeep: ShouldKeepTrue,
		AddObject:  AddObjectNil,
	})
	coup := &countingUpstream{
		fakeUpstream: up,
	}

	storage = New(cz, ca, newStdLogger())
	ctx = contexts.NewVhostContext(context.Background(), &types.VirtualHost{Upstream: coup}) // upstream is nil so that if used a panic will be produced
	makeAndCheckGetFullFile(t, ctx, storage, oid, expected)
	if err := storage.Close(); err != nil {
		t.Errorf("Storage.Close() shouldn't have returned error\n%s", err)
	}
	if coup.called != 0 {
		t.Errorf("Storage should've not called even once upstream but it did %d times", coup.called)
	}
}

func TestRestoreFromDisk2(t *testing.T) {
	t.Parallel()
	file_number := 64
	expected := "awesome"
	up, cz, _, _ := setup()
	ca := NewFakeCacheAlgorithm()
	ca.AddFakeReplies(types.ObjectIndex{}, FakeReplies{
		Lookup:     LookupFalse,
		ShouldKeep: ShouldKeepTrue,
		AddObject:  AddObjectNil,
	})

	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := 0; i < file_number; i++ {
		up.addFakeResponse(fmt.Sprintf("/path/%02d", i),
			fakeResponse{
				Status:       "200",
				ResponseTime: 0,
				Response:     expected,
			})
	}

	storage := New(cz, ca, newStdLogger())
	defer os.RemoveAll(storage.path)
	ctx := contexts.NewVhostContext(context.Background(), &types.VirtualHost{Upstream: up})

	for i := 0; i < file_number; i++ {
		oid := types.ObjectID{
			CacheKey: "1",
			Path:     fmt.Sprintf("/path/%02d", i),
		}

		makeAndCheckGetFullFile(t, ctx, storage, oid, expected)
	}

	if err := storage.Close(); err != nil {
		t.Errorf("Storage.Close() shouldn't have returned error\n%s", err)
	}

	coup := &countingUpstream{
		fakeUpstream: up,
	}

	var shouldKeepCount int32 = 0
	ca.AddFakeReplies(types.ObjectIndex{}, FakeReplies{
		Lookup: LookupTrue,
		ShouldKeep: func(o types.ObjectIndex) bool {
			atomic.AddInt32(&shouldKeepCount, 1)
			return true
		},
		AddObject: AddObjectNil,
	})

	storage = New(cz, ca, newStdLogger())
	if int(shouldKeepCount) != file_number {
		t.Errorf("Algorithm.ShouldKeep should've been called for each file a total of %d but was called %d times.", file_number, shouldKeepCount)
	}

	ctx = contexts.NewVhostContext(context.Background(), &types.VirtualHost{Upstream: coup})
	for i := 0; i < file_number; i++ {
		oid := types.ObjectID{
			CacheKey: "1",
			Path:     fmt.Sprintf("/path/%02d", i),
		}

		makeAndCheckGetFullFile(t, ctx, storage, oid, expected)
	}
	if err := storage.Close(); err != nil {
		t.Errorf("Storage.Close() shouldn't have returned error\n%s", err)
	}
	if coup.called != 0 {
		t.Errorf("Storage should've not called even once upstream but it did %d times", coup.called)
	}
}

func makeAndCheckGetFullFile(t *testing.T, ctx context.Context, storage *Disk, oid types.ObjectID, expected string) {
	resp, err := storage.GetFullFile(ctx, oid)
	if err != nil {
		t.Errorf("Got Error on a GetFullFile on closing storage:\n%s", err)
	}

	b, err := ioutil.ReadAll(resp)
	if err != nil {
		t.Errorf("Got Error on a ReadAll on an already closed storage:\n%s", err)
	}

	if string(b) != expected {
		t.Errorf("Expected read from GetFullFile was %s but got %s", expected, string(b))
	}
}
*/
