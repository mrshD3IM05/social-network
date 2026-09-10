import { redirect } from 'next/navigation'

// "/" sends you to the feed (the (main) layout sends you to /login if you are not logged in)
export default function Index() {
  redirect('/home')
}
