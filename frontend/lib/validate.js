// Every input rule of the app lives here, so the forms all check the same thing.
// The Go API checks it again — this only gives the user an answer right away.

export const LIMITS = {
  firstName: 50,
  lastName: 50,
  email: 254,
  nickname: { min: 4, max: 15 },
  password: { min: 8, max: 72 }, // bcrypt ignores anything past 72 characters
  aboutMe: 500,
  post: 1000,
  comment: 2000,        // same limit as the API (internal/service/commentsvc)
  message: 1000,
  search: 50,
  groupTitle: 100,       // same limit as the API (internal/service/groupsvc)
  groupDescription: 1000,
  minAge: 13,
  images: 3, // a post can carry at most 3 images
  imageBytes: 10 * 1024 * 1024, // 10 MB, like the API
  imageSide: 8000, // widest / tallest picture in pixels, like the API
  imageTypes: ['image/jpeg', 'image/png', 'image/gif'],
}

// value of the accept="" attribute of every file input
export const IMAGE_ACCEPT = LIMITS.imageTypes.join(',')

// same rules as internal/service/authsvc/service.go
const emailRegex = /^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$/
const nicknameRegex = /^[a-z0-9]{4,15}$/
const letterRegex = /[a-z]/

// Every check below answers with an error message, or '' when the value is fine.

// The {id} of a url like /chat/3. Next.js hands any segment to a [id] route, so
// it has to be checked before it goes into a request: /chat/abc and /chat/0 both
// reach the page. 0 means "not a real id".
//
// This is the same rule the API applies in common.PathID, which rejects anything
// under 1 with a 400 — so a bad id never even reaches the network.
export function parseId(value) {
  const id = Number(value)
  return Number.isInteger(id) && id >= 1 ? id : 0
}

// Required text with a maximum length: names, about me, a post, a message…
export function checkText(label, value, max, { required = true } = {}) {
  const text = (value || '').trim()
  if (required && !text) return `${label} is required.`
  if (text.length > max) return `${label} must be ${max} characters or less.`
  return ''
}

export function checkEmail(value) {
  const email = (value || '').trim().toLowerCase()
  if (!email) return 'Email is required.'
  if (email.length > LIMITS.email) return `Email must be ${LIMITS.email} characters or less.`
  if (!emailRegex.test(email)) return 'Enter a valid email address.'
  return ''
}

export function checkNickname(value) {
  const nickname = (value || '').trim().toLowerCase()
  if (!nickname) return '' // optional: the server makes one from the name
  if (!nicknameRegex.test(nickname) || !letterRegex.test(nickname)) {
    const { min, max } = LIMITS.nickname
    return `Nickname must be ${min}–${max} letters or numbers, with at least one letter.`
  }
  return ''
}

export function checkPassword(value) {
  const password = value || ''
  const { min, max } = LIMITS.password
  if (!password) return 'Password is required.'
  if (password.length < min) return `Password must be at least ${min} characters.`
  if (password.length > max) return `Password must be ${max} characters or less.`
  return ''
}

export function checkDateOfBirth(value) {
  if (!value) return 'Date of birth is required.'

  const birth = new Date(`${value}T00:00:00`)
  if (Number.isNaN(birth.getTime())) return 'Enter a valid date of birth.'

  const today = new Date()
  if (birth > today) return 'Date of birth cannot be in the future.'

  // full years between the two dates
  let age = today.getFullYear() - birth.getFullYear()
  if (
    today.getMonth() < birth.getMonth() ||
    (today.getMonth() === birth.getMonth() && today.getDate() < birth.getDate())
  ) {
    age--
  }

  if (age < LIMITS.minAge) return `You must be at least ${LIMITS.minAge} years old.`
  if (age > 120) return 'Enter a valid date of birth.'
  return ''
}

// Latest date of birth allowed, for the max="" attribute of the date input
export function maxBirthDate() {
  const date = new Date()
  date.setFullYear(date.getFullYear() - LIMITS.minAge)
  return date.toISOString().slice(0, 10)
}

// How many, which format, and how big — used by the post form
export function checkImages(files) {
  if (files.length > LIMITS.images) {
    return `You can add up to ${LIMITS.images} images per post.`
  }

  for (const file of files) {
    const error = checkImage(file)
    if (error) return error
  }
  return ''
}

// One image: format and size (avatars go through this one)
export function checkImage(file) {
  if (!LIMITS.imageTypes.includes(file.type) || file.size === 0) {
    return `"${file.name}" is not a JPEG, PNG or GIF image.`
  }
  if (file.size > LIMITS.imageBytes) {
    const mb = Math.round(LIMITS.imageBytes / (1024 * 1024))
    return `"${file.name}" is larger than ${mb} MB.`
  }
  return ''
}

// Width and height of a picture. Fails when the browser cannot open it,
// which also catches a file that only pretends to be an image.
async function readImageSize(file) {
  const url = URL.createObjectURL(file)
  try {
    const image = new Image()
    image.src = url
    await image.decode()
    return { width: image.naturalWidth, height: image.naturalHeight }
  } finally {
    URL.revokeObjectURL(url)
  }
}

// checkImage, then the picture is opened to prove it is a real image and to
// check its size in pixels. Used when the user picks a file.
export async function checkImageFile(file) {
  const error = checkImage(file)
  if (error) return error

  let size
  try {
    size = await readImageSize(file)
  } catch {
    return `"${file.name}" is not a valid image.`
  }
  if (size.width > LIMITS.imageSide || size.height > LIMITS.imageSide) {
    return `"${file.name}" is larger than ${LIMITS.imageSide}×${LIMITS.imageSide} pixels.`
  }
  return ''
}

// checkImages with the pixel check of checkImageFile
export async function checkImageFiles(files) {
  if (files.length > LIMITS.images) {
    return `You can add up to ${LIMITS.images} images per post.`
  }

  for (const file of files) {
    const error = await checkImageFile(file)
    if (error) return error
  }
  return ''
}

// First error of a list, or '' when every check passed
export function firstError(errors) {
  return errors.find(Boolean) || ''
}
