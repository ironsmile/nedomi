package mp4

import "io"

// uniContainer is universal container that takes all boxes
// can be embeded in other containers if the order of the contained
// boxes is not important
type uniContainer struct {
	boxes []Box
}

func decodeUniContainer(r io.Reader, size uint64) (*uniContainer, error) {
	l, err := DecodeContainer(r, size)
	if err != nil {
		return nil, err
	}
	u := &uniContainer{}
	for _, b := range l {
		u.addBox(b)
	}
	return u, nil
}

func (u *uniContainer) addBox(b Box) {
	u.boxes = append(u.boxes, b)
}

func (u *uniContainer) Encode(w io.Writer) error {
	for _, box := range u.boxes {
		if err := box.Encode(w); err != nil {
			return err
		}
	}

	return nil
}
