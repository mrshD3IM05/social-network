'use client'

import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { apiGet, apiPost } from '@/lib/api'
import { fetchPeople } from '@/lib/people'
import Modal from '@/components/Modal'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'
import CharCount from '@/components/CharCount'
import PostForm from '@/components/PostForm'
import PostCard from '@/components/PostCard'
import EventCard from '@/components/EventCard'
import EventFormModal from '@/components/EventFormModal'
import { LIMITS } from '@/lib/validate'

// One group: header bar with Group Info, posts (members), events (members),
// and a chat-style bottom bar with the Create Event action. The API hides
// groups you have no relation to (404) and gates posts/events to members.
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

  async function load() {
    try {
      const [detail, current] = await Promise.all([apiGet(`/groups/${id}`), apiGet('/me')])
      setGroup(detail)
      setMe(current)
    } catch (err) {
      if (err.status === 404) setNotFound(true)
      else setError(err.message)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  // Posts and events are only reachable for members; the API answers 403 for
  // outsiders and the sections simply stay empty instead of erroring.
  useEffect(() => {
    if (!isMember) {
      setPosts(null)
      setEvents(null)
      return
    }
    apiGet(`/groups/${id}/posts`).then(setPosts).catch(() => setPosts([]))
    apiGet(`/groups/${id}/events`).then(setEvents).catch(() => setEvents([]))
  }, [id, isMember])

  const loadPosts = useCallback(() => {
    apiGet(`/groups/${id}/posts`).then(setPosts).catch(err => setError(err.message))
  }, [id])

  async function respondJoinRequest(request, accept) {
    setError('')
    setNotice('')
    try {
      await apiPost(`/group-join-requests/${request.id}/${accept ? 'accept' : 'decline'}`)
      setNotice(accept ? `${request.first_name} is now a member.` : 'Request declined.')
      load()
    } catch (err) {
      setError(err.message)
    }
  }

  async function requestJoin() {
    setError('')
    setNotice('')
    try {
      await apiPost(`/groups/${id}/join-requests`)
      setNotice('Join request sent. The creator will review it.')
      load()
    } catch (err) {
      setError(err.message)
    }
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
        <button className="btn btn-light" onClick={() => setShowInfo(true)}>
          Group Info
        </button>
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

      {/* ------------------------------------------------ member view */}
      {isMember && (
        <>
          <p className="eyebrow section-label">Group posts</p>

          {posts === null && <p className="loading">Loading posts…</p>}
          {posts !== null && posts.length === 0 && (
            <div className="empty">
              <p className="empty-title">No posts yet</p>
              <p>Write the first one with the composer below.</p>
            </div>
          )}

          {posts?.map(post => (
            <PostCard key={post.id} post={post} myId={me.id} onDeleted={loadPosts} />
          ))}

          <p className="eyebrow section-label">Events</p>
          {events === null && <p className="loading">Loading events…</p>}
          {events !== null && events.length === 0 && (
            <div className="empty">
              <p className="empty-title">No events scheduled</p>
              <p>Create one with the + button below.</p>
            </div>
          )}
          {events?.map(event => (
            <EventCard
              key={event.id}
              event={event}
              onChanged={() => apiGet(`/groups/${id}/events`).then(setEvents).catch(() => {})}
            />
          ))}

          {/* chat-style composer at the bottom: the real post form (it
              publishes group posts) plus the Create Event action. */}
          <div className="group-composer">
            <PostForm groupId={id} onPosted={loadPosts} />
            <button
              className="btn btn-light group-composer-event"
              onClick={() => setShowEventForm(true)}
            >
              <Icon name="plus" size={16} /> Create Event
            </button>
          </div>
        </>
      )}

      {group.is_creator && (
        <JoinRequestsCard requests={group} onRespond={respondJoinRequest} onLoaded={setGroup} groupId={id} />
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

      {showInfo && (
        <InfoModal
          group={group}
          onClose={() => setShowInfo(false)}
          onInviteClick={() => {
            setShowInfo(false)
            setShowInvite(true)
          }}
        />
      )}

      {showEventForm && (
        <EventFormModal
          groupId={id}
          onClose={() => setShowEventForm(false)}
          onCreated={event => {
            setShowEventForm(false)
            setEvents(list => [...(list || []), event])
            setNotice(`Event "${event.title}" created. Members have been notified.`)
          }}
        />
      )}
    </>
  )
}

// Pending join requests, visible to the group creator only.
// Loaded from GET /groups/{id}/join-requests (403 for everyone else).
function JoinRequestsCard({ requests, onRespond, onLoaded, groupId }) {
  const [pending, setPending] = useState(null)

  useEffect(() => {
    apiGet(`/groups/${groupId}/join-requests`)
      .then(setPending)
      .catch(() => setPending([])) // not the creator per the API — hide the card
  }, [groupId, requests])

  if (pending === null) return null
  if (pending.length === 0) return null

  function respond(request, accept) {
    setPending(list => list.filter(r => r.id !== request.id))
    onRespond(request, accept)
  }

  return (
    <>
      <p className="eyebrow section-label">Join requests</p>
      <div className="card list">
        {pending.map(request => (
          <div key={request.id} className="list-item">
            <span className="list-icon"><Icon name="users" size={16} /></span>
            <span className="list-text">
              <strong>{request.first_name} {request.last_name} wants to join</strong>
              <small>@{request.nickname}</small>
            </span>
            <div className="invitation-actions">
              <button className="btn" onClick={() => respond(request, true)}>Accept</button>
              <button className="btn btn-light" onClick={() => respond(request, false)}>Decline</button>
            </div>
          </div>
        ))}
      </div>
    </>
  )
}

// Group Info dialog: everything GET /groups/{id} already provides — title,
// description, creator, members and the caller's own status.
function InfoModal({ group, onClose, onInviteClick }) {
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
        <div><dt>Members</dt><dd>{group.member_count}</dd></div>
        <div><dt>Created</dt><dd>{new Date(group.created_at).toLocaleDateString()}</dd></div>
        <div><dt>Your status</dt><dd>{group.is_creator ? 'Creator' : group.is_member ? 'Member' : 'Not a member'}</dd></div>
      </dl>

      <p className="eyebrow section-label">Members</p>
      <div className="info-members">
        {group.members.map(member => (
          <Link key={member.user_id} href={`/profile/${member.user_id}`} className="list-item">
            <Avatar user={member} size={36} />
            <span className="list-text">
              <strong>{member.first_name} {member.last_name}</strong>
              <small>@{member.nickname}</small>
            </span>
            {member.user_id === group.creator_id && <span className="chip chip-accent">Creator</span>}
          </Link>
        ))}
      </div>

      <div className="composer-bar">
        {onInviteClick && (
          <button className="btn btn-light" onClick={onInviteClick}>Invite people</button>
        )}
        <button className="btn" onClick={onClose}>Done</button>
      </div>
    </Modal>
  )
}

// Pick a person from your feed and invite them to the group.
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

  const shown = (people || []).filter(person =>
    !memberIds.has(person.id) &&
    `${person.first_name} ${person.last_name} ${person.nickname}`
      .toLowerCase()
      .includes(search.toLowerCase())
  )

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

      {people !== null && people.length === 0 && (
        <div className="empty">
          <p className="empty-title">No one to invite yet</p>
          <p>People show up here once their posts are in your feed.</p>
        </div>
      )}

      {people !== null && people.length > 0 && shown.length === 0 && (
        <p className="meta invite-none">No one matching — everyone is already a member or invited.</p>
      )}

      <div className="invite-list">
        {shown.map(person => (
          <div key={person.id} className="list-item">
            <Avatar user={person} size={40} />
            <span className="list-text">
              <strong>{person.first_name} {person.last_name}</strong>
              <small>@{person.nickname}</small>
            </span>
            {invited[person.id] ? (
              <span className="chip">Invited</span>
            ) : (
              <button className="btn btn-light btn-sm" onClick={() => invite(person)} disabled={busyId === person.id}>
                {busyId === person.id ? '…' : 'Invite'}
              </button>
            )}
          </div>
        ))}
      </div>

      <div className="composer-bar">
        <CharCount value={search} max={LIMITS.search} />
        <button className="btn btn-light" onClick={onClose}>Done</button>
      </div>
    </Modal>
  )
}
