export const enMessages = {
  common: {
    appName: "Falzo Travel",
    login: "Login",
    register: "Register",
    dashboard: "Dashboard",
    bookNow: "Book now",
    viewDetails: "View details",
  },
  home: {
    heroTitle: "Pick destination fast, see pricing clearly, book without friction",
    heroCta: "Explore 3D tours",
    searchTitle: "Find your right tour now",
    featuredTitle: "Featured tours",
    trustTitle: "Why travelers book with Falzo",
  },
  travel3d: {
    title: "Pick destinations on 3D globe and book faster",
    flowStep1: "Pick destination",
    flowStep2: "Review itinerary",
    flowStep3: "Book tour",
    detailsTitle: "Tour information based on selected destination",
  },
  auth: {
    loginSuccess: "Login successful",
    loginFailed: "Login failed",
    registerSuccess: "Registration successful",
    registerFailed: "Registration failed",
  },
  apiErrors: {
    unauthorized: "Incorrect email or password.",
    conflict: "This email is already registered.",
    server: "Server error. Please try again later.",
    unreachable: "Unable to connect to authentication API.",
    generic: "Something went wrong. Please try again.",
  },
} as const;

export type EnMessages = typeof enMessages;
