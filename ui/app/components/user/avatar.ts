// Generate UI Avatar URL with appropriate colors for light/dark mode
export function getAvatarUrl(name: string, darkMode: boolean) {
    // Default colors for light mode
    let backgroundColor = "f2f2f2"
    let foregroundColor = "333333"

    // Adjust colors for dark mode
    if (darkMode) {
      backgroundColor = "374151" // gray-700
      foregroundColor = "f3f4f6" // gray-100
    }

    // Encode the name for URL
    const encodedName = encodeURIComponent(name)

    return `https://ui-avatars.com/api/?name=${encodedName}&background=${backgroundColor}&color=${foregroundColor}&bold=true&format=svg`
  }