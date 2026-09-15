import PageHeader from '@/components/PageHeader'

const features = [
  { title: 'Create a group', text: 'Start a space around a topic and choose who joins.' },
  { title: 'Invite people', text: 'Invite your followers, or accept their requests.' },
  { title: 'Plan events', text: 'Set a date and see who is going.' },
  { title: 'Group chat', text: 'One conversation for every member, in real time.' },
]

// The API has no group endpoints yet, so this page only shows what is coming.
export default function GroupsPage() {
  return (
    <>
      <PageHeader label="Coming soon" title="Groups" subtitle="Communities, events and group chat are on the way." />

      <div className="features">
        {features.map((feature, index) => (
          <div key={feature.title} className="card feature">
            <span className="feature-number">0{index + 1}</span>
            <strong>{feature.title}</strong>
            <p>{feature.text}</p>
          </div>
        ))}
      </div>
    </>
  )
}
