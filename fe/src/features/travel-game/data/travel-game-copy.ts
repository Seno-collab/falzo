import type { SupportedLocale } from "@/i18n/messages";

export const destinationNodes = [
  {
    id: "phu-yen",
    imageId: "ly-son-coast",
    position: [-2.9, -0.1, 0.6] as const,
  },
  {
    id: "da-nang",
    imageId: "kyoto-lantern-night",
    position: [-1.2, 0.58, -0.35] as const,
  },
  {
    id: "mu-cang-chai",
    imageId: "mu-cang-chai-dawn",
    position: [0.45, 0.1, 0.1] as const,
  },
  {
    id: "hoi-an",
    imageId: "kyoto-lantern-night",
    position: [1.65, -0.34, -0.45] as const,
  },
  {
    id: "ly-son",
    imageId: "ly-son-coast",
    position: [2.85, 0.26, 0.35] as const,
  },
] as const;

export type DestinationNodeId = (typeof destinationNodes)[number]["id"];

export type TravelGameMission = {
  title: string;
  reward: string;
  description: string;
  status: string;
};

export type TravelGameCopy = {
  documentTitle: string;
  brand: string;
  navSubtitle: string;
  mobileTitle: string;
  headline: string;
  subHeadline: string;
  sampleCta: string;
  createCta: string;
  exploreCta: string;
  heroBadge: string;
  gameKicker: string;
  partyCta: string;
  proofItems: string[];
  landSelectorTitle: string;
  selectedLandLabel: string;
  startQuestCta: string;
  hudRank: string;
  xpProgressLabel: string;
  activeQuestLabel: string;
  claimRewardCta: string;
  claimedRewardLabel: string;
  rewardFeedbackTitle: string;
  rewardFeedbackDescription: string;
  dailyStreakLabel: string;
  dailyStreakValue: string;
  comboLabel: string;
  hudBadges: { label: string; value: string }[];
  gameplayFlow: string[];
  questTitle: string;
  questDescription: string;
  questStats: { label: string; value: string }[];
  missionsTitle: string;
  missionsDescription: string;
  missions: TravelGameMission[];
  leagueTitle: string;
  leagueDescription: string;
  leaderboard: { name: string; score: string; badge: string }[];
  pathTitle: string;
  pathLabel: string;
  pathDescription: string;
  pathSteps: string[];
  sceneLabel: string;
  sceneLoading: string;
  controlsLabel: string;
  controlsHint: string;
  moveUpLabel: string;
  moveDownLabel: string;
  moveLeftLabel: string;
  moveRightLabel: string;
  nodeLabel: string;
  destinations: Record<
    DestinationNodeId,
    {
      name: string;
      realm: string;
      mission: string;
      reward: string;
    }
  >;
  sectionLabel: string;
  sectionTitle: string;
  sectionDescription: string;
  detailLabel: string;
  samples: {
    title: string;
    imageId: string;
    province: string;
    duration: string;
    cost: string;
    stops: string;
    note: string;
  }[];
};

export const travelGameCopyByLocale: Record<SupportedLocale, TravelGameCopy> = {
  vi: {
    documentTitle: "Falzo.life - Travel game khám phá Việt Nam",
    brand: "Falzo.life",
    navSubtitle: "Travel game du lịch Việt Nam",
    mobileTitle: "Falzo game",
    headline: "Biến mỗi chuyến đi thành một cuộc chơi khám phá",
    subHeadline:
      "Săn địa điểm thật, upload ảnh thật, nhận XP và rủ bạn bè cùng leo bảng xếp hạng du lịch Việt Nam.",
    sampleCta: "Vào cuộc chơi",
    createCta: "Upload ảnh nhận XP",
    exploreCta: "Săn địa điểm",
    heroBadge: "Travel game challenge",
    gameKicker: "Mùa chơi 01",
    partyCta: "Rủ bạn cùng đi",
    proofItems: ["3 nhiệm vụ/ngày", "Upload ảnh thật", "Bảng xếp hạng"],
    landSelectorTitle: "Chọn vùng đất cổ tích",
    selectedLandLabel: "Vùng đất đang tới",
    startQuestCta: "Bắt đầu nhiệm vụ upload",
    hudRank: "Level 3 Explorer",
    xpProgressLabel: "Tiến độ lên rank",
    activeQuestLabel: "Nhiệm vụ đang theo",
    claimRewardCta: "Nhận XP nhiệm vụ",
    claimedRewardLabel: "Đã nhận XP",
    rewardFeedbackTitle: "XP đã cộng",
    rewardFeedbackDescription: "Chuỗi ngày và combo được cập nhật.",
    dailyStreakLabel: "Chuỗi ngày",
    dailyStreakValue: "5 ngày",
    comboLabel: "Combo",
    hudBadges: [
      { label: "Photo Hunter", value: "12 ảnh" },
      { label: "Route Builder", value: "4 route" },
      { label: "Local Scout", value: "8 mẹo" },
    ],
    gameplayFlow: [
      "Chọn điểm đến",
      "Bắt đầu nhiệm vụ",
      "Upload ảnh thật",
      "Nhận XP",
      "Leo rank",
    ],
    questTitle: "Cuộc chơi du lịch hôm nay",
    questDescription:
      "Chọn điểm đến trên bản đồ 3D, nhận nhiệm vụ, đăng ảnh kiểm chứng và mở khóa route cùng cộng đồng.",
    questStats: [
      { label: "người chơi tuần này", value: "2.4k" },
      { label: "ảnh đã upload", value: "860" },
      { label: "đội đang mở", value: "128" },
    ],
    missionsTitle: "Nhiệm vụ để bắt đầu ngay",
    missionsDescription:
      "Các thử thách ngắn biến việc xem travel content thành hành động cụ thể.",
    missions: [
      {
        title: "Săn ảnh thật",
        reward: "+120 XP",
        description: "Đăng một ảnh địa điểm có vị trí và ghi chú thực tế.",
        status: "Sẵn sàng",
      },
      {
        title: "Ghép route 24h",
        reward: "+180 XP",
        description: "Lưu 3 điểm đến và tạo một tuyến đi trong ngày.",
        status: "Nên làm hôm nay",
      },
      {
        title: "Rủ đồng đội",
        reward: "+90 XP",
        description: "Mời bạn bè bình chọn nơi đáng đi nhất cuối tuần.",
        status: "Nhiệm vụ nhanh",
      },
    ],
    leagueTitle: "Bảng xếp hạng khám phá",
    leagueDescription:
      "Hiển thị người đang đóng góp ảnh, route và mẹo hữu ích nhất.",
    leaderboard: [
      { name: "Mina", score: "8,420 XP", badge: "Trail Master" },
      { name: "Long", score: "7,980 XP", badge: "Local Scout" },
      { name: "Hana", score: "7,240 XP", badge: "Photo Hunter" },
    ],
    pathTitle: "Đường đua chuyến đi",
    pathLabel: "Cách chơi",
    pathDescription:
      "Người chơi đi qua từng chặng: khám phá, lưu điểm, đăng bằng chứng và mở khóa lịch trình.",
    pathSteps: [
      "Chọn bản đồ",
      "Săn địa điểm",
      "Upload ảnh",
      "Mở khóa route",
    ],
    sceneLabel: "Bản đồ 3D travel game",
    sceneLoading: "Đang mở bản đồ 3D",
    controlsLabel: "Điều khiển",
    controlsHint: "Di chuyển bằng WASD hoặc phím mũi tên để tới điểm đến.",
    moveUpLabel: "Lên",
    moveDownLabel: "Xuống",
    moveLeftLabel: "Trái",
    moveRightLabel: "Phải",
    nodeLabel: "Điểm chơi",
    destinations: {
      "phu-yen": {
        name: "Phú Yên",
        realm: "Vịnh ngọc biển xanh",
        mission: "Upload ảnh biển xanh hoặc ghềnh đá",
        reward: "+120 XP",
      },
      "da-nang": {
        name: "Đà Nẵng",
        realm: "Cổng mây Sơn Trà",
        mission: "Tạo route 24h từ biển đến phố cổ",
        reward: "+180 XP",
      },
      "mu-cang-chai": {
        name: "Mù Cang Chải",
        realm: "Thung lũng lúa vàng",
        mission: "Săn ảnh bình minh ruộng bậc thang",
        reward: "+160 XP",
      },
      "hoi-an": {
        name: "Hội An",
        realm: "Phố đèn lồng phép thuật",
        mission: "Chia sẻ mẹo đi phố cổ buổi tối",
        reward: "+90 XP",
      },
      "ly-son": {
        name: "Lý Sơn",
        realm: "Đảo núi lửa bí mật",
        mission: "Đăng gallery đảo và điểm check-in",
        reward: "+140 XP",
      },
    },
    sectionLabel: "Lịch trình mẫu",
    sectionTitle: "Bắt đầu từ những chuyến đi đã được sắp sẵn",
    sectionDescription:
      "Mỗi lịch trình gom địa điểm, thời gian đẹp, chi phí ước tính và lưu ý thực tế để bạn copy nhanh.",
    detailLabel: "Có trong lịch trình",
    samples: [
      {
        title: "Phú Yên 2 ngày săn biển xanh",
        imageId: "ly-son-coast",
        province: "Phú Yên",
        duration: "2 ngày",
        cost: "900k - 1.6tr",
        stops: "Hòn Yến, Gành Đá Đĩa, Mũi Điện",
        note: "Dễ đi xe máy, nên tránh ngày biển động.",
      },
      {
        title: "Mù Cang Chải mùa lúa 3 ngày",
        imageId: "mu-cang-chai-dawn",
        province: "Yên Bái",
        duration: "3 ngày",
        cost: "1.8tr - 3tr",
        stops: "Đèo Khau Phạ, La Pán Tẩn, Tú Lệ",
        note: "Đẹp nhất sáng sớm, cần đặt homestay trước.",
      },
      {
        title: "Đà Nẵng - Hội An chill 1 ngày",
        imageId: "kyoto-lantern-night",
        province: "Đà Nẵng",
        duration: "1 ngày",
        cost: "650k - 1.2tr",
        stops: "Sơn Trà, Mỹ Khê, phố cổ Hội An",
        note: "Phù hợp nhóm nhỏ và cặp đôi.",
      },
    ],
  },
  en: {
    documentTitle: "Falzo.life - Vietnam travel game",
    brand: "Falzo.life",
    navSubtitle: "Vietnam travel game",
    mobileTitle: "Falzo game",
    headline: "Turn every trip into a discovery game",
    subHeadline:
      "Hunt real places, upload real photos, earn XP, and invite friends to climb Vietnam's travel leaderboard.",
    sampleCta: "Join the game",
    createCta: "Upload for XP",
    exploreCta: "Hunt places",
    heroBadge: "Travel game challenge",
    gameKicker: "Season 01",
    partyCta: "Invite friends",
    proofItems: ["3 quests/day", "Real photo uploads", "Leaderboard"],
    landSelectorTitle: "Choose fairy-tale land",
    selectedLandLabel: "Current destination realm",
    startQuestCta: "Start upload quest",
    hudRank: "Level 3 Explorer",
    xpProgressLabel: "Rank progress",
    activeQuestLabel: "Active quest",
    claimRewardCta: "Claim quest XP",
    claimedRewardLabel: "XP claimed",
    rewardFeedbackTitle: "XP added",
    rewardFeedbackDescription: "Streak and combo have been updated.",
    dailyStreakLabel: "Daily streak",
    dailyStreakValue: "5 days",
    comboLabel: "Combo",
    hudBadges: [
      { label: "Photo Hunter", value: "12 photos" },
      { label: "Route Builder", value: "4 routes" },
      { label: "Local Scout", value: "8 tips" },
    ],
    gameplayFlow: [
      "Pick destination",
      "Start quest",
      "Upload proof",
      "Earn XP",
      "Rank up",
    ],
    questTitle: "Today's travel game",
    questDescription:
      "Pick a destination on the 3D map, take a quest, post proof photos, and unlock routes with the community.",
    questStats: [
      { label: "players this week", value: "2.4k" },
      { label: "photos uploaded", value: "860" },
      { label: "open teams", value: "128" },
    ],
    missionsTitle: "Missions to start now",
    missionsDescription:
      "Short challenges turn travel browsing into a reason to act.",
    missions: [
      {
        title: "Hunt real photos",
        reward: "+120 XP",
        description: "Post one geotagged place photo with practical notes.",
        status: "Ready",
      },
      {
        title: "Build a 24h route",
        reward: "+180 XP",
        description: "Save 3 places and turn them into a one-day route.",
        status: "Best today",
      },
      {
        title: "Bring a teammate",
        reward: "+90 XP",
        description: "Invite friends to vote for the best weekend stop.",
        status: "Quick quest",
      },
    ],
    leagueTitle: "Explorer leaderboard",
    leagueDescription:
      "Spotlights people contributing the most useful photos, routes, and local tips.",
    leaderboard: [
      { name: "Mina", score: "8,420 XP", badge: "Trail Master" },
      { name: "Long", score: "7,980 XP", badge: "Local Scout" },
      { name: "Hana", score: "7,240 XP", badge: "Photo Hunter" },
    ],
    pathTitle: "Trip race path",
    pathLabel: "How it works",
    pathDescription:
      "Players move through discovery, saving, proof posting, and itinerary unlocks.",
    pathSteps: ["Pick map", "Hunt places", "Upload photo", "Unlock route"],
    sceneLabel: "3D travel game map",
    sceneLoading: "Opening 3D map",
    controlsLabel: "Controls",
    controlsHint: "Move with WASD or arrow keys to reach a destination.",
    moveUpLabel: "Up",
    moveDownLabel: "Down",
    moveLeftLabel: "Left",
    moveRightLabel: "Right",
    nodeLabel: "Game stop",
    destinations: {
      "phu-yen": {
        name: "Phu Yen",
        realm: "Sapphire Coast Bay",
        mission: "Upload a blue-coast or stone-cliff photo",
        reward: "+120 XP",
      },
      "da-nang": {
        name: "Da Nang",
        realm: "Son Tra Cloud Gate",
        mission: "Build a 24h route from beach to old town",
        reward: "+180 XP",
      },
      "mu-cang-chai": {
        name: "Mu Cang Chai",
        realm: "Golden Terrace Valley",
        mission: "Hunt a sunrise terrace photo",
        reward: "+160 XP",
      },
      "hoi-an": {
        name: "Hoi An",
        realm: "Lantern Spell Town",
        mission: "Share one night-market local tip",
        reward: "+90 XP",
      },
      "ly-son": {
        name: "Ly Son",
        realm: "Secret Volcano Isle",
        mission: "Post an island gallery and check-in spot",
        reward: "+140 XP",
      },
    },
    sectionLabel: "Sample itineraries",
    sectionTitle: "Start from trips that are already structured",
    sectionDescription:
      "Each itinerary groups places, best timing, estimated budget, and practical notes so you can copy faster.",
    detailLabel: "Included stops",
    samples: [
      {
        title: "Phu Yen blue-coast 2-day plan",
        imageId: "ly-son-coast",
        province: "Phu Yen",
        duration: "2 days",
        cost: "900k - 1.6m VND",
        stops: "Hon Yen, Ganh Da Dia, Mui Dien",
        note: "Easy by motorbike, check sea conditions first.",
      },
      {
        title: "Mu Cang Chai rice-season 3 days",
        imageId: "mu-cang-chai-dawn",
        province: "Yen Bai",
        duration: "3 days",
        cost: "1.8m - 3m VND",
        stops: "Khau Pha Pass, La Pan Tan, Tu Le",
        note: "Best at sunrise, book homestays early.",
      },
      {
        title: "Da Nang - Hoi An 1-day chill route",
        imageId: "kyoto-lantern-night",
        province: "Da Nang",
        duration: "1 day",
        cost: "650k - 1.2m VND",
        stops: "Son Tra, My Khe, Hoi An old town",
        note: "Good for small groups and couples.",
      },
    ],
  },
};
