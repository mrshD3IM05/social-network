'use client'

import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { apiGet, apiPost } from '@/lib/api'
import { fetchPeople, searchPeople } from '@/lib/people'
import Modal from '@/components/Modal'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'
import CharCount from '@/components/CharCount'
import PersonRow from '@/components/PersonRow'
import PostForm from '@/components/PostForm'
import PostCard from '@/components/PostCard'
import EventCard from '@/components/EventCard'
import EventFormModal from '@/components/EventFormModal'
import { LIMITS } from '@/lib/validate'

// One group: header bar with the Invite and Group Info actions, a composer,
// events and posts (members only). The API hides groups you have no relation
// to (404) and gates posts/events to members.
export default function GroupDetailPage() {
  const { id } = useParams()
  const [me, setMe] = useState(null)
  const [group, setGroup] = useState(null)
  const [posts, setPosts] = useState(null)
  const [events, setEvents] = useState(null)
  const [notFound, setNotFound] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [showInvite, setShowInvite] = useState(false)
  const [showInfo, setShowInfo] = useState(false)
  const [showEventForm, setShowEventForm] = useState(false)

  const isMember = group?.is_member || group?.is_creator

  const load = useCallback(async () => {
    try {
      const [detail, current] = await Promise.all([apiGet(`/groups/${id}`), apiGet('/me')])
      setGroup(detail)
      setMe(current)
    } catch (err) {
      if (err.status === 404) setNotFound(true)
      else setError(err.message)
    }
  }, [id])

  // Posts and events sit behind the same membership check, so both the first
  // paint and every refresh go through these two helpers.
  const loadPosts = useCallback(
    () => apiGet(`/groups/${id}/posts`).then(setPosts).catch(() => setPosts([])),
    [id],
  )
  const loadEvents = useCallback(
    () => apiGet(`/groups/${id}/events`).then(setEvents).catch(() => setEvents([])),
    [id],
  )

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    if (!isMember) {
      setPosts(null)
      setEvents(null)
      return
    }
    loadPosts()
    loadEvents()
  }, [isMember, loadPosts, loadEvents])

  // Every action reports through the same notice/error pair and reloads the
  // group, so membership chips and counts stay in sync.
  async function run(action, message) {
    setError('')
    setNotice('')
    try {
      await action()
      setNotice(message)
      load()
    } catch (err) {
      setError(err.message)
    }
  }

  function respondJoinRequest(request, accept) {
    return run(
      () => apiPost(`/group-join-requests/${request.id}/${accept ? 'accept' : 'decline'}`),
      accept ? `${request.first_name} is now a member.` : 'Request declined.',
    )
  }

  function requestJoin() {
    return run(
      () => apiPost(`/groups/${id}/join-requests`),
      'Join request sent. The creator will review it.',
    )
  }

  if (notFound) {
    return (
      <div className="empty">
        <p className="empty-title">Group not found</p>
        <p>It does not exist, or you have no relation to it.</p>
        <Link href="/groups" className="btn">Back to groups</Link>
      </div>
    )
  }

  if (!group || !me) return <p className="loading">{error || 'Loading…'}</p>

  return (
    <>
      {/* ------------------------------------------------ header bar */}
      <div className="group-bar card">
        <Avatar user={{ avatar: null, first_name: group.title, last_name: '' }} size={44} />
        <div className="group-bar-title">
          <strong>{group.title}</strong>
          <small className="meta">
            {group.member_count} member{group.member_count === 1 ? '' : 's'}
            {group.is_creator && <> · you created this group</>}
          </small>
        </div>
        <div className="group-bar-actions">
          {isMember && (
            <button className="btn" onClick={() => setShowInvite(true)}>
              <Icon name="plus" size={16} /> Invite
            </button>
          )}
          <button className="btn btn-light" onClick={() => setShowInfo(true)}>Group Info</button>
        </div>
      </div>

      {error && <p className="error">{error}</p>}
      {notice && <p className="notice">{notice}</p>}

      {/* ------------------------------------------------ outsider view */}
      {!isMember && (
        <section className="card group-head">
          <div className="group-head-top">
            <span className="list-icon"><Icon name="grid" size={20} /></span>
            <div className="list-text">
              <strong>{group.description || 'No description.'}</strong>
              {group.creator && (
                <small>Created by {group.creator.first_name} {group.creator.last_name}</small>
              )}
            </div>
          </div>
          <div className="join-row">
            {group.pending_join ? (
              <span className="chip">Requested — waiting for the creator</span>
            ) : group.pending_invite ? (
              <p className="meta">You have been invited — accept it from the groups page.</p>
            ) : (
              <button className="btn" onClick={requestJoin}>Request to join</button>
            )}
          </div>
        </section>
      )}

      {/* The creator answers the queue first, then reads along. */}
      {group.is_creator && (
        <JoinRequestsCard groupId={id} refresh={group} onRespond={respondJoinRequest} />
      )}

      {/* ------------------------------------------------ member view */}
      {isMember && (
        <>
          {/* the real post form (it publishes group posts) plus the Create
              Event action, together at the top of the group. */}
          <p className="eyebrow section-label">Share with the group</p>
          <div className="group-composer">
            <PostForm groupId={id} onPosted={loadPosts} />
            <button
              className="btn btn-light group-composer-event"
              onClick={() => setShowEventForm(true)}
            >
              <Icon name="plus" size={16} /> Create Event
            </button>
          </div>

          <p className="eyebrow section-label">Events</p>
          {events === null && <p className="loading">Loading events…</p>}
          {events !== null && events.length === 0 && (
            <div className="empty">
              <p className="empty-title">No events scheduled</p>
              <p>Create one with the button above.</p>
            </div>
          )}
          {events?.map(event => (
            <EventCard key={event.id} event={event} onChanged={loadEvents} />
          ))}

          <p className="eyebrow section-label">Group posts</p>
          {posts === null && <p className="loading">Loading posts…</p>}
          {posts !== null && posts.length === 0 && (
            <div className="empty">
              <p className="empty-title">No posts yet</p>
              <p>Write the first one with the composer above.</p>
            </div>
          )}
          {posts?.map(post => (
            <PostCard key={post.id} post={post} myId={me.id} onDeleted={loadPosts} />
          ))}
        </>
      )}

      {showInvite && (
        <InviteModal
          groupId={id}
          memberIds={new Set(group.members.map(m => m.user_id))}
          onClose={() => setShowInvite(false)}
          onInvited={name => {
            setShowInvite(false)
            setNotice(`Invitation sent to ${name}.`)
          }}
        />
      )}

      {showInfo && <InfoModal group={group} onClose={() => setShowInfo(false)} />}

      {showEventForm && (
        <EventFormModal
          groupId={id}
          onClose={() => setShowEventForm(false)}
          onCreated={event => {
            setShowEventForm(false)
            loadEvents() // the API returns events in date order, so re-read
            setNotice(`Event "${event.title}" created. Members have been notified.`)
          }}
        />
      )}
    </>
  )
}

// Pending join requests, visible to the group creator only.
// Loaded from GET /groups/{id}/join-requests (403 for everyone else).
function JoinRequestsCard({ groupId, refresh, onRespond }) {
  const [pending, setPending] = useState(null)

  useEffect(() => {
    apiGet(`/groups/${groupId}/join-requests`)
      .then(setPending)
      .catch(() => setPending([])) // not the creator per the API — hide the card
  }, [groupId, refresh])

  if (!pending?.length) return null

  function respond(request, accept) {
    setPending(list => list.filter(r => r.id !== request.id))
    onRespond(request, accept)
  }

  return (
    <>
      <p className="eyebrow section-label">Join requests</p>
      <div className="card list">
        {pending.map(request => (
          <PersonRow key={request.id} person={request} size={40}>
            <div className="invitation-actions">
              <button className="btn btn-sm" onClick={() => respond(request, true)}>Accept</button>
              <button className="btn btn-light btn-sm" onClick={() => respond(request, false)}>Decline</button>
            </div>
          </PersonRow>
        ))}
      </div>
    </>
  )
}

// Group Info dialog: everything GET /groups/{id} already provides — title,
// description, creator, members and the caller's own status. Inviting lives
// in the header bar, not here.
function InfoModal({ group, onClose }) {
  return (
    <Modal title="Group Info" onClose={onClose}>
      <dl className="details">
        <div><dt>Title</dt><dd>{group.title}</dd></div>
        <div><dt>Description</dt><dd>{group.description || '—'}</dd></div>
        {group.creator && (
          <div>
            <dt>Creator</dt>
            <dd><Link href={`/profile/${group.creator.id}`}>{group.creator.first_name} {group.creator.last_name}</Link></dd>
          </div>
        )}
        <div><dt>Created</dt><dd>{new Date(group.created_at).toLocaleDateString()}</dd></div>
        <div><dt>Your status</dt><dd>{group.is_creator ? 'Creator' : group.is_member ? 'Member' : 'Not a member'}</dd></div>
      </dl>

      <p className="eyebrow section-label">Members · {group.member_count}</p>
      <div className="info-members">
        {group.members.map(member => (
          <PersonRow key={member.user_id} person={member} href={`/profile/${member.user_id}`} size={36}>
            {member.user_id === group.creator_id && <span className="chip chip-accent">Creator</span>}
          </PersonRow>
        ))}
      </div>

      <div className="composer-bar">
        <button className="btn" onClick={onClose}>Done</button>
      </div>
    </Modal>
  )
}

// Pick someone from the people directory (GET /users) and invite them.
// Members are filtered out; the API answers 409 for anyone already invited.
function InviteModal({ groupId, memberIds, onClose, onInvited }) {
  const [people, setPeople] = useState(null)
  const [search, setSearch] = useState('')
  const [invited, setInvited] = useState({}) // person id → true once the API confirms
  const [busyId, setBusyId] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    fetchPeople()
      .then(setPeople)
      .catch(err => setError(err.message))
  }, [])

  async function invite(person) {
    setError('')
    setBusyId(person.id)
    try {
      await apiPost(`/groups/${groupId}/invitations`, { user_id: person.id })
      setInvited(state => ({ ...state, [person.id]: true }))
      onInvited(person.first_name)
    } catch (err) {
      setError(err.message)
    }
    setBusyId(null)
  }

  const candidates = (people || []).filter(person => !memberIds.has(person.id))
  const shown = searchPeople(candidates, search)

  return (
    <Modal title="Invite people" onClose={onClose}>
      <div className="search">
        <Icon name="search" />
        <input
          placeholder="Search by name or nickname"
          value={search}
          maxLength={LIMITS.search}
          onChange={e => setSearch(e.target.value)}
          autoFocus
        />
      </div>

      {error && <p className="error">{error}</p>}

      {people === null && !error && <p className="loading">Loading…</p>}

      {people !== null && shown.length === 0 && (
        <div className="empty">
          <p className="empty-title">No one to invite</p>
          <p>
            {candidates.length === 0
              ? 'Everyone on the network is already a member.'
              : 'No one matches that search.'}
          </p>
        </div>
      )}

      <div className="invite-list">
        {shown.map(person => (
          <PersonRow key={person.id} person={person} size={40}>
            {invited[person.id] ? (
              <span className="chip">Invited</span>
            ) : (
              <button
                className="btn btn-light btn-sm"
                onClick={() => invite(person)}
                disabled={busyId === person.id}
              >
                {busyId === person.id ? '…' : 'Invite'}
              </button>
            )}
          </PersonRow>
        ))}
      </div>

      <div className="composer-bar">
        <CharCount value={search} max={LIMITS.search} />
        <button className="btn btn-light" onClick={onClose}>Done</button>
      </div>
    </Modal>
  )
}
