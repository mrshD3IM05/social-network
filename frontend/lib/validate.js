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
  message: 1000,
  search: 50,
  groupTitle: 100,       // same limit as the API (internal/service/groupsvc)
  groupDescription: 1000,
  minAge: 13,
  images: 3, // a post can carry at most 3 images
  imageBytes: 10 * 1024 * 1024, // 10 MB, like the API
  imageTypes: ['image/jpeg', 'image/png', 'image/gif'],
}

// value of the accept="" attribute of every file input
export const IMAGE_ACCEPT = LIMITS.imageTypes.join(',')

// same rules as internal/service/authsvc/service.go
const emailRegex = /^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$/
const nicknameRegex = /^[a-z0-9]{4,15}$/
const letterRegex = /[a-z]/

// Every check below answers with an error message, or '' when the value is fine.

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

// The subject marks the nickname as optional, so an empty value is fine;
// anything else still has to match the API's rule.
export function checkNickname(value, { required = false } = {}) {
  const nickname = (value || '').trim().toLowerCase()
  if (!nickname) return required ? 'Nickname is required.' : ''
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
  if (!LIMITS.imageTypes.includes(file.type)) {
    return `"${file.name}" is not a JPEG, PNG or GIF image.`
  }
  if (file.size > LIMITS.imageBytes) {
    const mb = Math.round(LIMITS.imageBytes / (1024 * 1024))
    return `"${file.name}" is larger than ${mb} MB.`
  }
  return ''
}

// First error of a list, or '' when every check passed
export function firstError(errors) {
  return errors.find(Boolean) || ''
}
