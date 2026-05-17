package jsonfile

import "time"

// lockRetryInterval is how often the backend re-polls a contested lock
// while waiting on a context. Short enough that the wait feels responsive
// to interactive use, long enough not to spin.
const lockRetryInterval = 25 * time.Millisecond
