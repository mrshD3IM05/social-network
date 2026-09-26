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

// One group: an identity header (who, what, how many, the actions) and one
// tab per thing the group holds — posts, events, members, and the creator's
// join requests. The API hides groups you have no relation to (404) and gates
// posts, events and members to members.
export default function GroupDetailPage() {
  const { id } = useParams()
  const [me, setMe] = useState(null)
  const [group, setGroup] = useState(null)
  const [posts, setPosts] = useState(null)
  const [events, setEvents] = useState(null)
  const [requests, setRequests] = useState([])
  const [tab, setTab] = useState('posts')
  const [notFound, setNotFound] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [showInvite, setShowInvite] = useState(false)
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

  // The three member-only lists. Each one is read through its own helper, so
  // the first paint and every refresh take the same path.
  const loadPosts = useCallback(
    () => apiGet(`/groups/${id}/posts`).then(setPosts).catch(() => setPosts([])),
    [id],
  )
  const loadEvents = useCallback(
    () => apiGet(`/groups/${id}/events`).then(setEvents).catch(() => setEvents([])),
    [id],
  )
  const loadRequests = useCallback(
    () => apiGet(`/groups/${id}/join-requests`).then(setRequests).catch(() => setRequests([])),
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

  useEffect(() => {
    if (group?.is_creator) loadRequests()
  }, [group?.is_creator, loadRequests])

  // Every action reports through the same notice/error pair and reloads the
  // group, so the counts and the status chip stay in sync.
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

  // Deleting a post only touches that one card: drop it from state instead
  // of refetching, so the page keeps its scroll position and the tab count
  // stays correct without a reload.
  const postDeleted = useCallback(
    postId => setPosts(list => (list ? list.filter(p => p.id !== postId) : list)),
    [],
  )

  function respondJoinRequest(request, accept) {
    setRequests(list => list.filter(r => r.id !== request.id))
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

  const tabs = [
    { key: 'posts', label: 'Posts', count: posts?.length },
    { key: 'events', label: 'Events', count: events?.length },
    { key: 'members', label: 'Members', count: group.member_count },
    ...(group.is_creator ? [{ key: 'requests', label: 'Requests', count: requests.length }] : []),
  ]

  return (
    <>
      {/* ------------------------------------------- identity header */}
      <header className="card group-hero">
        <div className="group-hero-top">
          <Avatar user={{ first_name: group.title, last_name: '' }} size={64} />
          <div className="group-hero-title">
            <h1>{group.title}</h1>
            <p className="meta">
              {group.member_count} member{group.member_count === 1 ? '' : 's'}
              {group.creator && <> · created by {group.creator.first_name} {group.creator.last_name}</>}
              {' · '}{new Date(group.created_at).toLocaleDateString()}
            </p>
          </div>
          {group.is_creator ? (
            <span className="chip chip-accent">Creator</span>
          ) : group.is_member ? (
            <span className="chip">Member</span>
          ) : null}
        </div>

        <p className="group-hero-desc">{group.description || 'No description.'}</p>

        <div className="group-hero-actions">
          <AvatarStack members={group.members} />
          {isMember ? (
            <button type="button" className="btn" onClick={() => setShowInvite(true)}>
              <Icon name="plus" size={16} /> Invite people
            </button>
          ) : group.pending_join ? (
            <p className="meta group-hero-note">Join request sent — waiting for the creator.</p>
          ) : group.pending_invite ? (
            <p className="meta group-hero-note">You are invited — answer it from the groups page.</p>
          ) : (
            <button type="button" className="btn" onClick={requestJoin}>Request to join</button>
          )}
        </div>
      </header>

      {error && <p className="error">{error}</p>}
      {notice && <p className="notice">{notice}</p>}

      {/* Outsiders stop here: the API serves nothing else to them. */}
      {!isMember ? (
        <Empty title="Members only">
          Posts, events and members open up once you join this group.
        </Empty>
      ) : (
        <>
          <nav className="tabs">
            {tabs.map(({ key, label, count }) => (
              <button
                key={key}
                className={`tab${tab === key ? ' active' : ''}`}
                onClick={() => setTab(key)}
              >
                {label}
                {count > 0 && <span className="tab-count">{count}</span>}
              </button>
            ))}
          </nav>

          {/* ------------------------------------------------- posts */}
          {tab === 'posts' && (
            <>
              <PostForm groupId={id} onPosted={loadPosts} />
              {posts === null && <p className="loading">Loading posts…</p>}
              {posts?.length === 0 && (
                <Empty title="No posts yet">Write the first one with the box above.</Empty>
              )}
              {posts?.map(post => (
                <PostCard
                  key={post.id}
                  post={post}
                  myId={me.id}
                  isGroupCreator={group.is_creator}
                  onDeleted={postDeleted}
                />
              ))}
            </>
          )}

          {/* ------------------------------------------------ events */}
          {tab === 'events' && (
            <>
              <div className="section-bar">
                <p className="eyebrow">Upcoming and past events</p>
                <button type="button" className="btn btn-sm" onClick={() => setShowEventForm(true)}>
                  <Icon name="plus" size={16} /> Create event
                </button>
              </div>
              {events === null && <p className="loading">Loading events…</p>}
              {events?.length === 0 && (
                <Empty title="No events scheduled">
                  Create one and every member gets notified.
                </Empty>
              )}
              {events?.map(event => (
                <EventCard key={event.id} event={event} onChanged={loadEvents} />
              ))}
            </>
          )}

          {/* ----------------------------------------------- members */}
          {tab === 'members' && (
            <div className="card list">
              {group.members.map(member => (
                <PersonRow key={member.user_id} person={member} href={`/profile/${member.user_id}`}>
                  {member.user_id === group.creator_id ? (
                    <span className="chip chip-accent">Creator</span>
                  ) : (
                    <Icon name="arrow" size={16} />
                  )}
                </PersonRow>
              ))}
            </div>
          )}

          {/* ---------------------------------------------- requests */}
          {tab === 'requests' && (
            requests.length === 0 ? (
              <Empty title="No pending requests">
                People asking to join this group land here.
              </Empty>
            ) : (
              <div className="card list">
                {requests.map(request => (
                  <PersonRow key={request.id} person={request} size={40}>
                    <div className="invitation-actions">
                      <button className="btn btn-sm" onClick={() => respondJoinRequest(request, true)}>Accept</button>
                      <button className="btn btn-light btn-sm" onClick={() => respondJoinRequest(request, false)}>Decline</button>
                    </div>
                  </PersonRow>
                ))}
              </div>
            )
          )}
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

// The faces of the group, so you see who is in it without opening anything.
function AvatarStack({ members, shown = 5 }) {
  const rest = members.length - shown
  return (
    <div className="avatar-stack">
      {members.slice(0, shown).map(member => (
        <Avatar key={member.user_id} user={member} size={32} />
      ))}
      {rest > 0 && <span className="avatar-stack-more">+{rest}</span>}
    </div>
  )
}

// The same empty state the group page shows in half a dozen places.
function Empty({ title, children }) {
  return (
    <div className="empty">
      <p className="empty-title">{title}</p>
      <p>{children}</p>
    </div>
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
        <Empty title="No one to invite">
          {candidates.length === 0
            ? 'Everyone on the network is already a member.'
            : 'No one matches that search.'}
        </Empty>
      )}

      <div className="invite-list">
        {shown.map(person => (
          <PersonRow key={person.id} person={person} size={40}>
            {invited[person.id] ? (
              <span className="chip">Invited</span>
            ) : (
              <button type="button" className="btn btn-light btn-sm" onClick={() => invite(person)} disabled={busyId === person.id}>
                {busyId === person.id ? '…' : 'Invite'}
              </button>
            )}
          </PersonRow>
        ))}
      </div>

      <div className="composer-bar">
        <CharCount value={search} max={LIMITS.search} />
        <button type="button" className="btn btn-light" onClick={onClose}>Done</button>
      </div>
    </Modal>
  )
}
