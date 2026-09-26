'use client'

import { useEffect, useState } from 'react'
import { apiGet, apiPost } from '@/lib/api'
import { fetchPeople } from '@/lib/people'
import { LIMITS, checkText } from '@/lib/validate'
import Modal from '@/components/Modal'
import Avatar from '@/components/Avatar'
import GroupCard from '@/components/GroupCard'
import Icon from '@/components/Icon'
import PageHeader from '@/components/PageHeader'
import CharCount from '@/components/CharCount'

// Groups hub: your invitations first, then the groups you belong to, then the
// ones left to discover — so the list you act on is never mixed with the rest.
export default function GroupsPage() {
  const [groups, setGroups] = useState(null)
  const [invitations, setInvitations] = useState([])
  const [people, setPeople] = useState({}) // user id → person, to name the inviters
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [joiningId, setJoiningId] = useState(null) // id of the group being joined

  function load() {
    apiGet('/groups')
      .then(setGroups)
      .catch(err => setError(err.message))
    apiGet('/group-invitations')
      .then(setInvitations)
      .catch(() => {}) // the inbox is secondary; the lists below still show
  }

  useEffect(() => {
    load()
    // the directory turns "user #3" into a name and a face on the invitations
    fetchPeople()
      .then(list => setPeople(Object.fromEntries(list.map(person => [person.id, person]))))
      .catch(() => {})
  }, [])

  async function respondInvitation(id, accept) {
    setError('')
    try {
      await apiPost(`/group-invitations/${id}/${accept ? 'accept' : 'decline'}`)
      setInvitations(list => list.filter(inv => inv.id !== id))
      load() // membership changed → the group moves to "Your groups"
    } catch (err) {
      setError(err.message)
    }
  }

  async function join(group) {
    setError('')
    setJoiningId(group.id)
    try {
      await apiPost(`/groups/${group.id}/join-requests`)
      setGroups(list =>
        list.map(g => (g.id === group.id ? { ...g, pending_join: true } : g)),
      )
    } catch (err) {
      setError(err.message)
    }
    setJoiningId(null)
  }

  const mine = groups?.filter(group => group.is_member || group.is_creator) || []
  const others = groups?.filter(group => !group.is_member && !group.is_creator) || []

  return (
    <>
      <PageHeader label="Communities" title="Groups" subtitle="Find your people, or start a space of your own." />

      <div className="section-bar">
        <p className="eyebrow">
          {groups === null ? 'Loading…' : `${mine.length} joined · ${others.length} to discover`}
        </p>
        <button type="button" className="btn" onClick={() => setShowCreate(true)}>
          <Icon name="plus" size={16} /> Create a group
        </button>
      </div>

      {error && <p className="error">{error}</p>}

      {invitations.length > 0 && (
        <section className="card invitations">
          <h2>Group invitations</h2>
          {invitations.map(inv => {
            const from = people[inv.from_user_id]
            return (
              <div key={inv.id} className="list-item">
                {from
                  ? <Avatar user={from} size={40} />
                  : <span className="list-icon"><Icon name="users" size={16} /></span>}
                <span className="list-text">
                  <strong>You are invited to join “{inv.group_title}”</strong>
                  <small>{from ? `${from.first_name} ${from.last_name} invited you` : 'You have a pending invitation'}</small>
                </span>
                <div className="invitation-actions">
                  <button className="btn btn-sm" onClick={() => respondInvitation(inv.id, true)}>Accept</button>
                  <button className="btn btn-light btn-sm" onClick={() => respondInvitation(inv.id, false)}>Decline</button>
                </div>
              </div>
            )
          })}
        </section>
      )}

      {groups !== null && (
        <>
          <GroupSection
            label="Your groups"
            groups={mine}
            empty="You have not joined a group yet. Pick one below, or create your own."
            onJoin={join}
            joiningId={joiningId}
          />
          <GroupSection
            label="Discover"
            groups={others}
            empty="Nothing left to discover — you are in every group."
            onJoin={join}
            joiningId={joiningId}
          />
        </>
      )}

      {showCreate && (
        <CreateGroupModal
          onClose={() => setShowCreate(false)}
          onCreated={() => {
            setShowCreate(false)
            load() // the new group shows up under "Your groups" (as creator)
          }}
        />
      )}
    </>
  )
}

// One titled list of groups, or a one-line reason why it is empty.
function GroupSection({ label, groups, empty, onJoin, joiningId }) {
  return (
    <>
      <p className="eyebrow section-label">{label}</p>
      {groups.length === 0 ? (
        <p className="meta section-empty">{empty}</p>
      ) : (
        <div className="card list">
          {groups.map(group => (
            <GroupCard key={group.id} group={group} onJoin={onJoin} joining={joiningId === group.id} />
          ))}
        </div>
      )}
    </>
  )
}

// Create form in a modal. The API answers 201 with the new group.
function CreateGroupModal({ onClose, onCreated }) {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const titleError = checkText('Title', title, LIMITS.groupTitle)
  const descriptionError = checkText('Description', description, LIMITS.groupDescription, { required: false })

  async function handleSubmit(e) {
    e.preventDefault()
    const problem = titleError || descriptionError
    if (problem) {
      setError(problem)
      return
    }
    setError('')
    setLoading(true)
    try {
      const group = await apiPost('/groups', { title: title.trim(), description: description.trim() })
      onCreated(group)
    } catch (err) {
      setError(err.message)
      setLoading(false)
    }
  }

  return (
    <Modal title="Create a group" onClose={onClose}>
      <form onSubmit={handleSubmit} noValidate>
        <label>Title</label>
        <input
          value={title}
          maxLength={LIMITS.groupTitle}
          className={error && titleError ? 'invalid' : undefined}
          onChange={e => setTitle(e.target.value)}
          autoFocus
        />

        <label>Description <small>optional</small></label>
        <textarea
          rows={3}
          value={description}
          maxLength={LIMITS.groupDescription}
          className={error && descriptionError ? 'invalid' : undefined}
          onChange={e => setDescription(e.target.value)}
        />

        <div className="composer-bar">
          <CharCount value={description} max={LIMITS.groupDescription} />
          <button className="btn" disabled={loading || Boolean(titleError) || Boolean(descriptionError)}>
            {loading ? 'Creating…' : 'Create group'}
          </button>
        </div>

        {error && <p className="error">{error}</p>}
      </form>
    </Modal>
  )
}
