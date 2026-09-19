'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { apiGet, apiPost } from '@/lib/api'
import { fetchPeople } from '@/lib/people'
import Modal from '@/components/Modal'
import Avatar from '@/components/Avatar'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'
import CharCount from '@/components/CharCount'
import { LIMITS } from '@/lib/validate'

// One group: info, member list, invitations (members) and join-request
// management (creator). The API hides groups you have no relation to (404).
export default function GroupDetailPage() {
  const { id } = useParams()
  const [me, setMe] = useState(null)
  const [group, setGroup] = useState(null)
  const [notFound, setNotFound] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [showInvite, setShowInvite] = useState(false)

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

  const canInvite = group.is_member || group.is_creator

  return (
    <>
      <PageHeader label="Group" title={group.title} subtitle={group.description || 'No description.'} />

      {error && <p className="error">{error}</p>}
      {notice && <p className="notice">{notice}</p>}

      <section className="card group-head">
        <div className="group-head-top">
          <span className="list-icon"><Icon name="grid" size={20} /></span>
          <div className="list-text">
            <strong>
              {group.member_count} member{group.member_count === 1 ? '' : 's'}
              {group.creator && <> · created by {group.creator.first_name} {group.creator.last_name}</>}
            </strong>
            {group.is_creator && <small>You created this group.</small>}
          </div>
          {canInvite && (
            <button className="btn" onClick={() => setShowInvite(true)}>
              <Icon name="plus" size={16} /> Invite
            </button>
          )}
        </div>

        {!group.is_member && !group.is_creator && (
          <div className="join-row">
            {group.pending_join ? (
              <span className="chip">Requested — waiting for the creator</span>
            ) : group.pending_invite ? (
              <p className="meta">You have been invited — accept it from the groups page.</p>
            ) : (
              <button className="btn" onClick={requestJoin}>Request to join</button>
            )}
          </div>
        )}
      </section>

      {group.is_creator && (
        <JoinRequestsCard requests={group} onRespond={respondJoinRequest} onLoaded={setGroup} groupId={id} />
      )}

      <p className="eyebrow section-label">Members</p>
      <div className="card list">
        {group.members.map(member => (
          <Link key={member.user_id} href={`/profile/${member.user_id}`} className="list-item">
            <Avatar user={member} size={44} />
            <span className="list-text">
              <strong>{member.first_name} {member.last_name}</strong>
              <small>@{member.nickname}</small>
            </span>
            {member.user_id === group.creator_id && <span className="chip chip-accent">Creator</span>}
          </Link>
        ))}
      </div>

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
