// The title block at the top of every page
export default function PageHeader({ label, title, subtitle }) {
  return (
    <header className="page-header">
      {label && <p className="eyebrow">{label}</p>}
      <h1>{title}</h1>
      {subtitle && <p className="subtitle">{subtitle}</p>}
    </header>
  )
}
