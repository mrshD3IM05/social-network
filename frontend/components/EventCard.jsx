'use client'

import { useState } from 'react'
import { apiPost } from '@/lib/api'
import Icon from './Icon'

const choiceLabels = { going: 'Going', not_going: 'Not going' }

// One group event: title, description, date/time, going counts and the
// Going / Not going buttons. The answer updates in place — the API answers
// with the new counts and the stored choice.
export default function EventCard({ event, onChanged }) {
  const [going, setGoing] = useState(event.going_count)
  const [notGoing, setNotGoing] = useState(event.not_going_count)
  const [myChoice, setMyChoice] = useState(event.my_choice || '')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')

  const when = new Date(event.date_time)
  const day = when.toLocaleDateString(undefined, { weekday: 'short', year: 'numeric', month: 'short', day: 'numeric' })
  const time = when.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
  const past = when.getTime() < Date.now()

  async function respond(choice) {
    setError('')
    setBusy(true)
    // optimistic: flip the answer immediately, correct it if the API refuses
    const previous = { going, notGoing, myChoice }
    const addedGoing = choice === 'going'
    const removedGoing = myChoice === 'going' && choice !== 'going'
    const removedNot = myChoice === 'not_going' && choice !== 'not_going'
    setGoing(count => count + (addedGoing ? 1 : 0) - (removedGoing ? 1 : 0))
    setNotGoing(count => count + (choice === 'not_going' ? 1 : 0) - (removedNot ? 1 : 0))
    setMyChoice(choice)
    try {
      const result = await apiPost(`/events/${event.id}/response`, { choice })
      setGoing(result.going_count)
      setNotGoing(result.not_going_count)
      setMyChoice(result.my_choice)
      if (onChanged) onChanged(result)
    } catch (err) {
      setGoing(previous.going)
      setNotGoing(previous.notGoing)
      setMyChoice(previous.myChoice)
      setError(err.message)
    }
    setBusy(false)
  }

  return (
    <article className="card event">
      <header className="event-head">
        <span className="list-icon"><Icon name="bell" size={18} /></span>
        <div className="event-who">
          <h3>{event.title}</h3>
          <small className="meta">
            by {event.creator_first_name} {event.creator_last_name}
          </small>
        </div>
        <div className="event-when">
          <strong>{day}</strong>
          <span className="meta">{time}{past ? ' · past' : ''}</span>
        </div>
      </header>

      {event.description && <p className="event-description">{event.description}</p>}

      <footer className="event-footer">
        <span className="meta">
          Going: <strong>{going}</strong> · Not going: <strong>{notGoing}</strong>
          {myChoice && <> · you: <strong>{choiceLabels[myChoice]}</strong></>}
        </span>
        <div className="event-actions">
          <button
            className={myChoice === 'going' ? 'btn btn-sm' : 'btn btn-light btn-sm'}
            disabled={busy}
            onClick={() => respond('going')}
          >
            Going
          </button>
          <button
            className={myChoice === 'not_going' ? 'btn btn-sm' : 'btn btn-light btn-sm'}
            disabled={busy}
            onClick={() => respond('not_going')}
          >
            Not going
          </button>
        </div>
      </footer>

      {error && <p className="error">{error}</p>}
    </article>
  )
}
