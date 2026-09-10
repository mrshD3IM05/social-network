# Frontend Implementation Guide

Owner: Person B (frontend)
Depends on: Backend API shapes from docs/API.md (can mock initially)
Blocks: nothing (can start with WSProvider + standalone components)

---

## Step 1 — WebSocketProvider

**New file:** `frontend/social-network-fn/src/app/providers/WebSocketProvider.js`

A React Context that manages a single shared WebSocket connection for the entire app.

```jsx
"use client";

import { createContext, useContext, useEffect, useRef, useState, useCallback } from "react";

const WebSocketContext = createContext(null);

export function useWebSocket() {
  return useContext(WebSocketContext);
}

export function WebSocketProvider({ children }) {
  const [connected, setConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState(null);
  const socketRef = useRef(null);
  const reconnectTimeout = useRef(null);
  const reconnectDelay = useRef(1000);
  const listenersRef = useRef(new Map());

  const connect = useCallback(() => {
    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    const ws = new WebSocket(`${protocol}://${window.location.host}/api/v1/ws`);
    socketRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      reconnectDelay.current = 1000; // reset backoff
    };

    ws.onclose = () => {
      setConnected(false);
      // exponential backoff reconnect
      reconnectTimeout.current = setTimeout(() => {
        reconnectDelay.current = Math.min(reconnectDelay.current * 2, 30000);
        connect();
      }, reconnectDelay.current);
    };

    ws.onmessage = (event) => {
      const payload = JSON.parse(event.data);
      setLastMessage(payload);
      // notify type-specific listeners
      const callbacks = listenersRef.current.get(payload.type);
      if (callbacks) {
        callbacks.forEach((cb) => cb(payload));
      }
    };
  }, []);

  useEffect(() => {
    connect();
    return () => {
      clearTimeout(reconnectTimeout.current);
      socketRef.current?.close();
    };
  }, [connect]);

  const send = useCallback((type, payload = {}) => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify({ type, ...payload }));
    }
  }, []);

  const subscribe = useCallback((type, callback) => {
    if (!listenersRef.current.has(type)) {
      listenersRef.current.set(type, new Set());
    }
    listenersRef.current.get(type).add(callback);
    return () => listenersRef.current.get(type)?.delete(callback);
  }, []);

  return (
    <WebSocketContext.Provider value={{ connected, lastMessage, send, subscribe }}>
      {children}
    </WebSocketContext.Provider>
  );
}
```

### Usage in any component:
```jsx
const { connected, send, subscribe } = useWebSocket();

// Send a message
send("message", { to_user_id: 5, content: "Hello!" });

// Listen for incoming messages
useEffect(() => {
  return subscribe("message", (payload) => {
    console.log("New message:", payload.message);
  });
}, [subscribe]);
```

---

## Step 2 — UnreadProvider

**New file:** `frontend/social-network-fn/src/app/providers/UnreadProvider.js`

Tracks the notification unread count (client-side) and updates it via WebSocket events.
Per-conversation message unread badges are intentionally not tracked.

```jsx
"use client";

import { createContext, useContext, useEffect, useState, useCallback } from "react";
import { useWebSocket } from "./WebSocketProvider";
import { api } from "../components/SocialShell";

const UnreadContext = createContext(null);

export function useUnread() {
  return useContext(UnreadContext);
}

export function UnreadProvider({ children }) {
  const [unreadNotifications, setUnreadNotifications] = useState(0);
  const { subscribe } = useWebSocket();

  // Fetch initial notification count on mount
  useEffect(() => {
    api("/notifications?unread=true")
      .then((data) => setUnreadNotifications(data.unread_count || 0))
      .catch(() => {});
  }, []);

  // Listen for new notifications
  useEffect(() => {
    return subscribe("notification", () => {
      setUnreadNotifications((prev) => prev + 1);
    });
  }, [subscribe]);

  const markNotificationsRead = useCallback(() => {
    setUnreadNotifications(0);
    api("/notifications/read-all", { method: "POST" }).catch(() => {});
  }, []);

  return (
    <UnreadContext.Provider value={{
      unreadNotifications,
      markNotificationsRead,
    }}>
      {children}
    </UnreadContext.Provider>
  );
}
```

---

## Step 3 — ChatLayout (conversation list sidebar)

**New file:** `frontend/social-network-fn/src/app/components/ChatLayout.js`

A sidebar that shows all conversations (private + group) with last message preview.

```jsx
"use client";

import { useEffect, useState } from "react";
import { api, displayName } from "./SocialShell";
import styles from "../page.module.css";

export default function ChatLayout({ activeConversation, onSelect }) {
  const [conversations, setConversations] = useState({ private: [], groups: [] });
  const [search, setSearch] = useState("");

  useEffect(() => {
    api("/messages/conversations").then(setConversations).catch(() => {});
  }, []);

  const filtered = {
    private: conversations.private?.filter((c) =>
      displayName(c.user).toLowerCase().includes(search.toLowerCase())
    ) || [],
    groups: conversations.groups?.filter((c) =>
      c.group.title.toLowerCase().includes(search.toLowerCase())
    ) || [],
  };

  function select(conversation) {
    onSelect(conversation);
  }

  return (
    <aside className={styles.chatSidebar}>
      <div className={styles.chatSidebarHeader}>
        <h3>Messages</h3>
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search conversations..."
          aria-label="Search conversations"
        />
      </div>
      <div className={styles.conversationList}>
        {/* Private conversations */}
        {filtered.private.map((conv) => (
          <button
            key={`p-${conv.user.id}`}
            className={`${styles.conversationItem} ${
              activeConversation?.type === "private" && activeConversation?.id === conv.user.id
                ? styles.activeConversation
                : ""
            }`}
            onClick={() => select({ type: "private", id: conv.user.id, user: conv.user })}
          >
            <div className={styles.conversationAvatar}>
              {/* Avatar component here */}
            </div>
            <div className={styles.conversationInfo}>
              <strong>{displayName(conv.user)}</strong>
              <span className={styles.conversationPreview}>
                {conv.last_message?.content || "No messages yet"}
              </span>
            </div>
            <div className={styles.conversationMeta}>
              <span className={styles.conversationTime}>
                {conv.last_message?.created_at
                  ? new Date(conv.last_message.created_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
                  : ""}
              </span>
            </div>
          </button>
        ))}

        {/* Group conversations */}
        {filtered.groups.map((conv) => (
          <button
            key={`g-${conv.group.id}`}
            className={`${styles.conversationItem} ${
              activeConversation?.type === "group" && activeConversation?.id === conv.group.id
                ? styles.activeConversation
                : ""
            }`}
            onClick={() => select({ type: "group", id: conv.group.id, group: conv.group })}
          >
            {/* Similar layout, group icon instead of avatar */}
            <div className={styles.conversationInfo}>
              <strong>{conv.group.title}</strong>
              <span className={styles.conversationPreview}>
                {conv.last_message?.content || "No messages yet"}
              </span>
            </div>
            <div className={styles.conversationMeta}>
              <span className={styles.conversationTime}>
                {conv.last_message?.created_at
                  ? new Date(conv.last_message.created_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
                  : ""}
              </span>
            </div>
          </button>
        ))}

        {filtered.private.length === 0 && filtered.groups.length === 0 && (
          <div className={styles.empty}>No conversations yet. Start chatting!</div>
        )}
      </div>
    </aside>
  );
}
```

---

## Step 4 — MessageThread

**New file:** `frontend/social-network-fn/src/app/components/MessageThread.js`

Displays messages between two users or in a group, with input and emoji picker.

Image attachments: send images with the message by posting to `POST /files` with `message_id` after the message is created (max 3 images, jpeg/png/gif). Wire this through a shared `fileUrl(id)` helper and render attachments inside the bubble.

```jsx
"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { api, displayName } from "./SocialShell";
import { useWebSocket } from "../providers/WebSocketProvider";
import EmojiPicker from "./EmojiPicker";
import styles from "../page.module.css";

export default function MessageThread({ conversation, me }) {
  const [messages, setMessages] = useState([]);
  const [draft, setDraft] = useState("");
  const [hasMore, setHasMore] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [showEmoji, setShowEmoji] = useState(false);
  const messagesEndRef = useRef(null);
  const messagesContainerRef = useRef(null);
  const { send, subscribe } = useWebSocket();

  const isGroup = conversation.type === "group";
  const endpoint = isGroup
    ? `/messages/group/${conversation.id}`
    : `/messages/${conversation.id}`;

  // Load initial messages
  useEffect(() => {
    setMessages([]);
    setHasMore(true);
    api(`${endpoint}?limit=50`).then((data) => {
      setMessages(data.messages.reverse()); // API returns newest first, we want oldest first
      setHasMore(data.has_more);
    }).catch(() => {});
  }, [endpoint]);

  // Listen for incoming messages
  useEffect(() => {
    return subscribe("message", (payload) => {
      const msg = payload.message;
      const isRelevant = isGroup
        ? msg.group_id === conversation.id
        : (msg.from_user_id === conversation.id && msg.to_user_id === me.id) ||
          (msg.from_user_id === me.id && msg.to_user_id === conversation.id);
      if (isRelevant) {
        setMessages((prev) => [...prev, msg]);
      }
    });
  }, [subscribe, conversation, me, isGroup]);

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // Load older messages (infinite scroll upward)
  const loadMore = useCallback(async () => {
    if (loadingMore || !hasMore || messages.length === 0) return;
    setLoadingMore(true);
    const oldestId = messages[0].id;
    try {
      const data = await api(`${endpoint}?before=${oldestId}&limit=50`);
      setMessages((prev) => [...data.messages.reverse(), ...prev]);
      setHasMore(data.has_more);
    } catch {}
    setLoadingMore(false);
  }, [endpoint, messages, hasMore, loadingMore]);

  // Handle scroll to top for loading more
  const handleScroll = useCallback((e) => {
    if (e.target.scrollTop === 0 && hasMore) {
      loadMore();
    }
  }, [hasMore, loadMore]);

  // Send message
  function sendMessage(e) {
    e.preventDefault();
    if (!draft.trim()) return;
    const payload = isGroup
      ? { group_id: conversation.id }
      : { to_user_id: conversation.id };
    send("message", { ...payload, content: draft.trim() });
    setDraft("");
  }

  return (
    <section className={styles.messageThread}>
      <div className={styles.threadHeader}>
        <h2>{isGroup ? conversation.group.title : displayName(conversation.user)}</h2>
      </div>

      <div
        className={styles.messageList}
        ref={messagesContainerRef}
        onScroll={handleScroll}
      >
        {loadingMore && <div className={styles.loadingOlder}>Loading older messages...</div>}
        {!hasMore && messages.length > 0 && <div className={styles.noOlder}>Beginning of conversation</div>}

        {messages.map((msg, i) => {
          const isMine = msg.from_user_id === me.id;
          return (
            <div
              key={`${msg.id}-${i}`}
              className={`${styles.messageBubble} ${isMine ? styles.sent : styles.received}`}
            >
              {!isMine && isGroup && (
                <span className={styles.senderName}>
                  {msg.from_user_id === me.id ? "You" : `User #${msg.from_user_id}`}
                </span>
              )}
              <p>{msg.content}</p>
              <span className={styles.messageTime}>
                {new Date(msg.created_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
              </span>
            </div>
          );
        })}
        <div ref={messagesEndRef} />
      </div>

      <form className={styles.chatInput} onSubmit={sendMessage}>
        <button type="button" className={styles.emojiButton} onClick={() => setShowEmoji(!showEmoji)}>
          😀
        </button>
        {showEmoji && (
          <EmojiPicker
            onSelect={(emoji) => setDraft((prev) => prev + emoji)}
            onClose={() => setShowEmoji(false)}
          />
        )}
        <input
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          placeholder="Type a message..."
          aria-label="Message input"
        />
        <button type="submit" className={styles.primaryButton} disabled={!draft.trim()}>
          Send
        </button>
      </form>
    </section>
  );
}
```

---

## Step 5 — EmojiPicker

**New file:** `frontend/social-network-fn/src/app/components/EmojiPicker.js`

A standalone emoji picker component.

```jsx
"use client";

import { useState, useRef, useEffect } from "react";
import styles from "../page.module.css";

const EMOJI_CATEGORIES = {
  "Smileys": ["😀","😃","😄","😁","😆","😅","🤣","😂","🙂","😊","😇","🥰","😍","🤩","😘","😗","😚","😙","🥲","😋","😛","😜","🤪","😝","🤑","🤗","🤭","🫢","🫣","🤫","🤔","🫡","🤐","🤨","😐","😑","😶","🫥","😏","😒","🙄","😬","🤥","😌","😔","😪","🤤","😴","😷","🤒","🤕","🤢","🤮","🥵","🥶","🥴","😵","🤯","🤠","🥳","🥸","😎","🤓","🧐"],
  "Gestures": ["👋","🤚","🖐️","✋","🖖","🫱","🫲","🫳","🫴","👌","🤌","🤏","✌️","🤞","🫰","🤟","🤘","🤙","👈","👉","👆","🖕","👇","☝️","🫵","👍","👎","✊","👊","🤛","🤜","👏","🙌","🫶","👐","🤲","🤝","🙏"],
  "Hearts": ["❤️","🧡","💛","💚","💙","💜","🖤","🤍","🤎","💔","❤️‍🔥","❤️‍🩹","❣️","💕","💞","💓","💗","💖","💘","💝"],
  "Objects": ["🎉","🎊","🎈","🎁","🎂","🍰","🍩","🍪","☕","🍵","🥤","🍺","🍷","🥂","音乐","🎵","🎶","🔔","📱","💻","⌨️","🖥️","📷","📸","🔑","🗝️","💡","📚","📝"],
};

export default function EmojiPicker({ onSelect, onClose }) {
  const [search, setSearch] = useState("");
  const ref = useRef(null);

  // Close on outside click
  useEffect(() => {
    function handleClick(e) {
      if (ref.current && !ref.current.contains(e.target)) onClose();
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [onClose]);

  const filtered = {};
  for (const [category, emojis] of Object.entries(EMOJI_CATEGORIES)) {
    const match = emojis.filter((e) => !search || e.includes(search));
    if (match.length) filtered[category] = match;
  }

  return (
    <div className={styles.emojiPicker} ref={ref}>
      <input
        className={styles.emojiSearch}
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Search emoji..."
        aria-label="Search emoji"
      />
      <div className={styles.emojiGrid}>
        {Object.entries(filtered).map(([category, emojis]) => (
          <div key={category}>
            <p className={styles.emojiCategoryLabel}>{category}</p>
            <div className={styles.emojiRow}>
              {emojis.map((emoji) => (
                <button
                  key={emoji}
                  className={styles.emojiButton}
                  onClick={() => { onSelect(emoji); onClose(); }}
                  type="button"
                >
                  {emoji}
                </button>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
```

---

## Step 6 — Rewrite messages/page.js

**File:** `frontend/social-network-fn/src/app/messages/page.js`

Replace the current content entirely. The new page uses `ChatLayout` + `MessageThread`.

```jsx
"use client";

import { useState } from "react";
import SocialShell from "../components/SocialShell";
import ChatLayout from "../components/ChatLayout";
import MessageThread from "../components/MessageThread";
import { useWebSocket } from "../providers/WebSocketProvider";
import { api } from "../components/SocialShell";
import styles from "../page.module.css";

export default function MessagesPage() {
  const [activeConversation, setActiveConversation] = useState(null);
  const [me, setMe] = useState(null);

  // Fetch current user for MessageThread
  useState(() => { api("/me").then(setMe).catch(() => {}); }, []);

  return (
    <SocialShell active="messages" eyebrow="COMMON GROUND / MESSAGES" title="Conversations with a pulse">
      <div className={styles.chatLayout}>
        <ChatLayout
          activeConversation={activeConversation}
          onSelect={setActiveConversation}
        />
        {activeConversation && me ? (
          <MessageThread conversation={activeConversation} me={me} />
        ) : (
          <section className={`${styles.panel} ${styles.chatPanel}`}>
            <div className={styles.empty}>
              <h3>Select a conversation</h3>
              <p>Choose someone from the list to start chatting.</p>
            </div>
          </section>
        )}
      </div>
    </SocialShell>
  );
}
```

---

## Step 7 — Rewrite notifications/page.js

**File:** `frontend/social-network-fn/src/app/notifications/page.js`

Replace with full notification history + real-time + action buttons.

```jsx
"use client";

import { useEffect, useState } from "react";
import SocialShell, { api, displayName } from "../components/SocialShell";
import { useWebSocket } from "../providers/WebSocketProvider";
import { useUnread } from "../providers/UnreadProvider";
import styles from "../page.module.css";

export default function NotificationsPage() {
  const [items, setItems] = useState([]);
  const [hasMore, setHasMore] = useState(true);
  const { connected, subscribe } = useWebSocket();
  const { markNotificationsRead } = useUnread();

  // Load history
  useEffect(() => {
    api("/notifications?limit=50").then((data) => {
      setItems(data.notifications || []);
      setHasMore(data.has_more);
      markNotificationsRead();
    }).catch(() => {});
  }, []);

  // Listen for real-time notifications
  useEffect(() => {
    return subscribe("notification", (payload) => {
      setItems((prev) => [payload.notification, ...prev]);
    });
  }, [subscribe]);

  // Load more (older)
  function loadMore() {
    if (!hasMore || items.length === 0) return;
    const oldestId = items[items.length - 1].id;
    api(`/notifications?before=${oldestId}&limit=50`).then((data) => {
      setItems((prev) => [...prev, ...(data.notifications || [])]);
      setHasMore(data.has_more);
    }).catch(() => {});
  }

  // Accept/decline actions
  async function handleFollowAction(notificationId, action) {
    await api(`/follow-requests/${notificationId}/${action}`, { method: "POST" });
    setItems((prev) => prev.filter((n) => n.id !== notificationId));
  }

  async function handleGroupAction(notificationId, action) {
    // TODO: implement when group endpoints exist
  }

  return (
    <SocialShell active="notifications" eyebrow="COMMON GROUND / NOTIFICATIONS" title="A little closer to the action">
      <div className={styles.pageIntro}>
        <p>Follow requests, group invitations, join requests, and events appear here.</p>
        <span className={styles.statusPill}>{connected ? "Live" : "Connecting..."}</span>
      </div>
      <section className={styles.listPanel}>
        {items.length ? (
          items.map((item) => (
            <article className={`${styles.notification} ${!item.read ? styles.unreadNotification : ""}`} key={item.id}>
              <span className={styles.notificationDot} />
              <div>
                <strong>{item.type.replace(/_/g, " ")}</strong>
                <p>{item.content}</p>
                <span>{new Date(item.created_at).toLocaleString()}</span>
              </div>
              <div className={styles.notificationActions}>
                {item.type === "follow_request" && (
                  <>
                    <button onClick={() => handleFollowAction(item.actor_id, "accept")}>Accept</button>
                    <button onClick={() => handleFollowAction(item.actor_id, "decline")}>Decline</button>
                  </>
                )}
                {item.type === "group_invite" && (
                  <>
                    <button onClick={() => handleGroupAction(item.id, "accept")}>Accept</button>
                    <button onClick={() => handleGroupAction(item.id, "decline")}>Decline</button>
                  </>
                )}
              </div>
            </article>
          ))
        ) : (
          <div className={styles.empty}>No notifications yet.</div>
        )}
        {hasMore && items.length > 0 && (
          <button className={styles.ghostButton} onClick={loadMore}>Load more</button>
        )}
      </section>
    </SocialShell>
  );
}
```

---

## Step 8 — Update SocialShell.js (badges)

**File:** `frontend/social-network-fn/src/app/components/SocialShell.js`

Add an unread count badge to the Notifications nav link (messages have no unread badges).

Changes:
1. Import `useUnread` from providers
2. Add a badge count next to the "Notifications" nav link

```jsx
// In the nav links section, replace:
// {links.map(([href, label, key]) => <Link ...>{label}</Link>)}

// With something like:
import { useUnread } from "../providers/UnreadProvider";

// Inside the component:
const { unreadNotifications } = useUnread();

const links = [
  ["/feed", "Feed", "feed"],
  ["/profile/" + me.id, "Profile", "profile"],
  ["/groups", "Groups", "groups"],
  ["/messages", "Messages", "messages"],
  ["/notifications", "Notifications", "notifications", unreadNotifications],
  ["/settings", "Settings", "settings"],
];

// In the render:
{links.map(([href, label, key, count]) => (
  <Link className={active === key ? styles.activeNav : ""} href={href} key={key}>
    {label}
    {count > 0 && <span className={styles.navBadge}>{count}</span>}
  </Link>
))}
```

---

## Step 9 — Update layout.js (providers)

**File:** `frontend/social-network-fn/src/app/layout.js`

Wrap the app with WebSocketProvider and UnreadProvider.

```jsx
import { WebSocketProvider } from "./providers/WebSocketProvider";
import { UnreadProvider } from "./providers/UnreadProvider";

// In the body or a client component wrapper:
<WebSocketProvider>
  <UnreadProvider>
    {children}
  </UnreadProvider>
</WebSocketProvider>
```

Note: Since layout.js may be a server component, you may need a client-side wrapper component that includes both providers. Create `src/app/providers/AppProviders.js` if needed.

---

## Step 10 — CSS additions

**File:** `frontend/social-network-fn/src/app/page.module.css`

Add styles for all new components. Key additions:

```css
/* Chat Layout */
.chatLayout {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: 1px;
  height: calc(100vh - 120px);
  background: var(--border);
}

/* Conversation Sidebar */
.chatSidebar {
  background: var(--bg);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.chatSidebarHeader {
  padding: 1rem;
  border-bottom: 1px solid var(--border);
}
.chatSidebarHeader input {
  width: 100%;
  margin-top: 0.5rem;
  padding: 0.5rem;
}
.conversationList {
  flex: 1;
  overflow-y: auto;
}
.conversationItem {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  width: 100%;
  border: none;
  background: none;
  text-align: left;
  cursor: pointer;
  border-bottom: 1px solid var(--border);
}
.conversationItem:hover, .activeConversation {
  background: var(--hover);
}
.conversationInfo {
  flex: 1;
  min-width: 0;
}
.conversationPreview {
  display: block;
  font-size: 0.85rem;
  color: var(--muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.conversationMeta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.25rem;
}
.conversationTime {
  font-size: 0.75rem;
  color: var(--muted);
}

/* Message Thread */
.messageThread {
  display: flex;
  flex-direction: column;
  background: var(--bg);
}
.threadHeader {
  padding: 1rem;
  border-bottom: 1px solid var(--border);
}
.messageList {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.messageBubble {
  max-width: 70%;
  padding: 0.5rem 0.75rem;
  border-radius: 1rem;
  font-size: 0.9rem;
}
.sent {
  align-self: flex-end;
  background: var(--primary);
  color: white;
  border-bottom-right-radius: 0.25rem;
}
.received {
  align-self: flex-start;
  background: var(--surface);
  border-bottom-left-radius: 0.25rem;
}
.senderName {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--primary);
  display: block;
  margin-bottom: 0.15rem;
}
.messageTime {
  font-size: 0.7rem;
  opacity: 0.7;
  display: block;
  text-align: right;
}

/* Chat Input */
.chatInput {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--border);
}
.chatInput input {
  flex: 1;
}

/* Emoji Picker */
.emojiPicker {
  position: absolute;
  bottom: 100%;
  left: 0;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 0.5rem;
  padding: 0.5rem;
  width: 300px;
  max-height: 300px;
  overflow-y: auto;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 10;
}
.emojiSearch {
  width: 100%;
  margin-bottom: 0.5rem;
  padding: 0.4rem;
}
.emojiCategoryLabel {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--muted);
  margin: 0.5rem 0 0.25rem;
}
.emojiRow {
  display: flex;
  flex-wrap: wrap;
  gap: 0.15rem;
}
.emojiPicker .emojiButton {
  background: none;
  border: none;
  font-size: 1.25rem;
  padding: 0.25rem;
  cursor: pointer;
  border-radius: 0.25rem;
}
.emojiPicker .emojiButton:hover {
  background: var(--hover);
}

/* Notifications */
.unreadNotification {
  border-left: 3px solid var(--primary);
  background: var(--hover);
}
.notificationActions {
  display: flex;
  gap: 0.5rem;
}
.notificationActions button {
  padding: 0.25rem 0.75rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: var(--bg);
  cursor: pointer;
}

/* Nav badges */
.navBadge {
  background: var(--primary);
  color: white;
  border-radius: 999px;
  padding: 0.1rem 0.35rem;
  font-size: 0.7rem;
  margin-left: 0.25rem;
  vertical-align: super;
}

/* Responsive */
@media (max-width: 768px) {
  .chatLayout {
    grid-template-columns: 1fr;
  }
  .chatSidebar.hidden-mobile {
    display: none;
  }
  .messageBubble {
    max-width: 85%;
  }
}
```

---

## File Checklist

| Action | File |
|--------|------|
| New | `src/app/providers/WebSocketProvider.js` |
| New | `src/app/providers/UnreadProvider.js` |
| New | `src/app/components/ChatLayout.js` |
| New | `src/app/components/MessageThread.js` |
| New | `src/app/components/EmojiPicker.js` |
| Rewrite | `src/app/messages/page.js` |
| Rewrite | `src/app/notifications/page.js` |
| Modify | `src/app/components/SocialShell.js` |
| Modify | `src/app/layout.js` (or create AppProviders wrapper) |
| Modify | `src/app/page.module.css` |
