import { apiGet } from '@/lib/api'

// Derive the list of known people from the posts in the feed.
// (The API has no "list users" endpoint — the same pattern lives in
// people/page.jsx and chat/page.jsx, which now import this helper.)
export async function fetchPeople() {
  const me = await apiGet('/me')
  const posts = await apiGet('/posts')

  const found = {}
  for (const post of posts) {
    if (post.author_id !== me.id) {
      found[post.author_id] = {
        id: post.author_id,
        first_name: post.author_first_name,
        last_name: post.author_last_name,
        nickname: post.author_nickname,
        avatar: post.author_avatar,
      }
    }
  }
  return Object.values(found)
}
