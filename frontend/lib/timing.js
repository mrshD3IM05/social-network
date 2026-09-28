import { useEffect, useMemo, useRef, useState } from 'react'

// Throttle with a leading AND a trailing call:
//   - the first call runs right away (leading)
//   - calls during the next `wait` ms are not dropped: the last one runs once
//     the wait is over (trailing)
// A plain "once per window" throttle only has the leading call, so whatever
// happens at the end of a burst is lost. This one always ends on the latest call.
export function throttle(fn, wait) {
  let last = 0 // when fn last ran
  let timer = null
  let pendingArgs = null

  function run() {
    last = Date.now()
    timer = null
    const args = pendingArgs
    pendingArgs = null
    fn(...args)
  }

  function throttled(...args) {
    pendingArgs = args
    const remaining = wait - (Date.now() - last)
    if (remaining <= 0) {
      clearTimeout(timer)
      run()
    } else if (!timer) {
      timer = setTimeout(run, remaining)
    }
  }

  // drop a trailing call that has not run yet
  throttled.cancel = () => {
    clearTimeout(timer)
    timer = null
    pendingArgs = null
  }

  return throttled
}

// throttle() for components: always calls the latest fn (so it sees fresh
// state), keeps the same throttled function between renders, and cancels the
// pending call when the component goes away.
export function useThrottle(fn, wait) {
  const fnRef = useRef(fn)
  useEffect(() => {
    fnRef.current = fn
  })

  const throttled = useMemo(() => throttle((...args) => fnRef.current(...args), wait), [wait])
  useEffect(() => () => throttled.cancel(), [throttled])

  return throttled
}

// Debounce a value: gives back `value` only once it has stopped changing for
// `delay` ms. Keep the input bound to the real value so typing stays instant,
// and use the debounced one for the work that follows.
export function useDebouncedValue(value, delay) {
  const [debounced, setDebounced] = useState(value)

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(timer)
  }, [value, delay])

  return debounced
}
