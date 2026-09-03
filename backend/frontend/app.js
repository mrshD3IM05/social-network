const API_BASE = '/api/v1';

const state = { me: null, posts: [] };
const elements = {
  authView: document.querySelector('#auth-view'),
  appView: document.querySelector('#app-view'),
  authMessage: document.querySelector('#auth-message'),
  loginForm: document.querySelector('#login-form'),
  registerForm: document.querySelector('#register-form'),
  meCard: document.querySelector('#me-card'),
  welcome: document.querySelector('#welcome'),
  privacyBadge: document.querySelector('#privacy-badge'),
  feed: document.querySelector('#feed'),
  toast: document.querySelector('#toast'),
  profileResult: document.querySelector('#profile-result'),
  lightbox: document.querySelector('#lightbox'),
  lightboxImage: document.querySelector('#lightbox-image'),
};

const formBody = (form) => new URLSearchParams(new FormData(form));

const AVATAR_FILE_ID = /^[0-9a-f]{32}$/i;

const fileURL = (fileID) => `${API_BASE}/fs/${encodeURIComponent(fileID)}`;

function avatarFileID(value) {
  value = String(value ?? '');
  return AVATAR_FILE_ID.test(value) ? value : '';
}

function avatarSrc(value) {
  const fileID = avatarFileID(value);
  if (fileID) return `${fileURL(fileID)}/thumb`;
  if (/^https?:\/\//i.test(String(value ?? ''))) return String(value);
  return '';
}

async function api(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    ...options,
  });
  const text = await response.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch { data = text; }
  if (!response.ok) throw new Error(typeof data === 'string' ? data : `Request failed (${response.status})`);
  return data;
}

function showToast(message) {
  elements.toast.textContent = message;
  elements.toast.classList.add('show');
  window.setTimeout(() => elements.toast.classList.remove('show'), 3000);
}

function setAuthMessage(message) { elements.authMessage.textContent = message; }

function openPreview(fileID) {
  elements.lightboxImage.src = fileURL(fileID);
  elements.lightbox.classList.remove('hidden');
}

function closePreview() {
  if (elements.lightbox.classList.contains('hidden')) return;
  elements.lightboxImage.src = '';
  elements.lightbox.classList.add('hidden');
}

function renderMe() {
  const user = state.me;
  elements.welcome.textContent = `Welcome, ${user.first_name}`;
  elements.privacyBadge.textContent = user.private ? 'PRIVATE' : 'PUBLIC';
  const avatar = avatarSrc(user.avatar);
  const avatarFile = avatarFileID(user.avatar);
  const avatarHTML = avatar ? `<img class="avatar previewable" ${avatarFile ? `data-file-id="${escapeHtml(avatarFile)}"` : ''} src="${escapeHtml(avatar)}" alt="Avatar">` : '';
  elements.meCard.innerHTML = `${avatarHTML}<strong class="profile-name">${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)}</strong><span class="profile-meta">@${escapeHtml(user.nickname || 'no nickname')}</span><span class="profile-meta">${escapeHtml(user.email)}</span>`;
}

function authorLabel(post) {
  if (post.author_id === state.me.id) return 'You';
  const name = [post.author_first_name, post.author_last_name].map((part) => String(part ?? '').trim()).filter(Boolean).join(' ');
  if (name) return name;
  if (post.author_nickname) return `@${post.author_nickname}`;
  return `User #${post.author_id}`;
}

function renderPosts() {
  if (!state.posts?.length) {
    elements.feed.innerHTML = '<div class="empty">No visible posts yet. Publish the first one.</div>';
    return;
  }
  elements.feed.innerHTML = state.posts.map((post) => {
    const authorAvatar = avatarSrc(post.author_avatar);
    const authorAvatarFile = avatarFileID(post.author_avatar);
    return `
    <article class="post">
      <div class="post-header"><div class="post-author">${authorAvatar ? `<img class="avatar feed-avatar previewable" ${authorAvatarFile ? `data-file-id="${escapeHtml(authorAvatarFile)}"` : ''} src="${escapeHtml(authorAvatar)}" alt="">` : ''}<div><strong>${escapeHtml(authorLabel(post))}</strong><div class="post-meta">${escapeHtml(post.privacy)} · ${formatDate(post.created_at)}</div></div></div>${post.author_id === state.me.id ? `<div class="post-actions"><button class="small-button" data-edit-post="${post.id}">Edit</button><button class="small-button" data-delete-post="${post.id}">Delete</button></div>` : ''}</div>
      <div class="post-content">${escapeHtml(post.content)}</div>
      ${post.images?.length ? `<div class="post-images">${post.images.map((fileID) => `<img class="previewable" data-file-id="${escapeHtml(fileID)}" src="${fileURL(fileID)}/thumb" alt="Attached image" loading="lazy">`).join('')}</div>` : ''}
      <div class="post-reactions"><button class="reaction${post.my_reaction === 'like' ? ' active' : ''}" data-react-post="${post.id}" data-reaction="like">&#128077; ${post.likes}</button><button class="reaction${post.my_reaction === 'dislike' ? ' active' : ''}" data-react-post="${post.id}" data-reaction="dislike">&#128078; ${post.dislikes}</button></div>
    </article>`;
  }).join('');
}

function formatDate(value) { return value ? new Date(value).toLocaleString() : ''; }
function escapeHtml(value) { return String(value ?? '').replace(/[&<>'"]/g, (character) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' }[character])); }

async function loadWorkspace() {
  try {
    state.me = await api('/me');
    state.posts = await api('/posts');
    renderMe(); renderPosts();
    elements.authView.classList.add('hidden');
    elements.appView.classList.remove('hidden');
    document.querySelector('#api-status').textContent = 'API connected /api/v1';
  } catch (error) {
    elements.authView.classList.remove('hidden');
    elements.appView.classList.add('hidden');
    document.querySelector('#api-status').textContent = 'API /api/v1';
  }
}

async function submitAuth(event, endpoint, form) {
  event.preventDefault(); setAuthMessage('Working...');
  try {
    await api(endpoint, { method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body: formBody(form) });
    await loadWorkspace(); setAuthMessage('');
  } catch (error) { setAuthMessage(error.message); }
}

document.querySelectorAll('[data-auth-tab]').forEach((tab) => tab.addEventListener('click', () => {
  const register = tab.dataset.authTab === 'register';
  document.querySelectorAll('.tab').forEach((item) => item.classList.toggle('active', item === tab));
  elements.loginForm.classList.toggle('hidden', register);
  elements.registerForm.classList.toggle('hidden', !register);
  setAuthMessage('');
}));

elements.loginForm.addEventListener('submit', (event) => submitAuth(event, '/login', elements.loginForm));
elements.registerForm.addEventListener('submit', (event) => submitAuth(event, '/register', elements.registerForm));

document.addEventListener('click', (event) => {
  const previewable = event.target.closest('[data-file-id]');
  if (previewable) openPreview(previewable.dataset.fileId);
});

elements.lightbox.addEventListener('click', closePreview);

window.addEventListener('keydown', (event) => { if (event.key === 'Escape') closePreview(); });

document.querySelector('#logout-button').addEventListener('click', async () => {
  try { await api('/logout', { method: 'POST' }); } catch (error) { showToast(error.message); }
  state.me = null; state.posts = []; elements.appView.classList.add('hidden'); elements.authView.classList.remove('hidden');
});

document.querySelector('#refresh-button').addEventListener('click', async () => {
  try { state.posts = await api('/posts'); renderPosts(); showToast('Feed refreshed'); } catch (error) { showToast(error.message); }
});

document.querySelector('#avatar-input').addEventListener('change', async (event) => {
  const file = event.target.files[0];
  event.target.value = '';
  if (!file) return;
  const upload = new FormData();
  upload.append('avatar', file);
  try {
    state.me = await api('/avatar', { method: 'POST', body: upload });
    renderMe(); showToast('Avatar updated');
  } catch (error) { showToast(error.message); }
});

document.querySelector('#post-form').addEventListener('submit', async (event) => {
  event.preventDefault(); const form = event.currentTarget;
  try {
    const post = await api('/posts', { method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body: new URLSearchParams({ content: form.elements.content.value, privacy: form.elements.privacy.value }) });
    const files = [...form.elements.files.files];
    if (files.length) {
      const upload = new FormData(); files.forEach((file) => upload.append('files', file)); upload.append('post_id', post.id);
      await api('/files', { method: 'POST', body: upload });
    }
    form.reset(); state.posts = await api('/posts'); renderPosts(); showToast(files.length ? `Post and ${files.length} image${files.length === 1 ? '' : 's'} published` : 'Post published');
  } catch (error) { showToast(error.message); }
});

document.querySelector('#profile-form').addEventListener('submit', async (event) => {
  event.preventDefault(); const id = new FormData(event.currentTarget).get('id');
  try {
    const profile = await api(`/user/${id}`);
    elements.profileResult.innerHTML = `<strong>${escapeHtml(profile.first_name)} ${escapeHtml(profile.last_name)}</strong><br><span>${escapeHtml(profile.nickname || 'No nickname')} · ${profile.private ? 'Private profile' : 'Public profile'}</span><br><button class="small-button" data-follow-user="${profile.id}">Follow</button>`;
  } catch (error) { elements.profileResult.textContent = error.message; }
});

elements.profileResult.addEventListener('click', async (event) => {
  const userID = event.target.dataset.followUser;
  if (!userID) return;
  try { await api(`/users/${userID}/follow`, { method: 'POST' }); showToast('Follow request sent'); } catch (error) { showToast(error.message); }
});

elements.feed.addEventListener('click', async (event) => {
  const reactID = event.target.dataset.reactPost;
  if (reactID) {
    // the API toggles: sending the reaction already on the post removes it
    try {
      const summary = await api(`/posts/${reactID}/reactions`, { method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body: new URLSearchParams({ reaction: event.target.dataset.reaction }) });
      const post = state.posts.find((item) => String(item.id) === reactID);
      if (post) { post.likes = summary.likes; post.dislikes = summary.dislikes; post.my_reaction = summary.my_reaction; renderPosts(); }
    } catch (error) { showToast(error.message); }
    return;
  }
  const postID = event.target.dataset.deletePost;
  if (postID) {
    try { await api(`/posts/${postID}`, { method: 'DELETE' }); state.posts = await api('/posts'); renderPosts(); showToast('Post deleted'); } catch (error) { showToast(error.message); }
  }
  const editID = event.target.dataset.editPost;
  if (editID) {
    const content = window.prompt('Update post content');
    if (content === null) return;
    try { await api(`/posts/${editID}`, { method: 'PUT', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body: new URLSearchParams({ content, privacy: 'public' }) }); state.posts = await api('/posts'); renderPosts(); showToast('Post updated'); } catch (error) { showToast(error.message); }
  }
});

loadWorkspace();
