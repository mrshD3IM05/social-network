// "120/1000" under a text field — turns red once the limit is passed.
export default function CharCount({ value, max }) {
  const used = value.trim().length
  return (
    <span className={used > max ? 'counter over' : 'counter'}>
      {used}/{max}
    </span>
  )
}
