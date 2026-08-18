export type RelationshipStatus =
  | "NONE"
  | "FRIENDS"
  | "INCOMING_REQUEST"
  | "OUTGOING_REQUEST";

export type FriendRequestStatus = "PENDING" | "ACCEPTED" | "REJECTED" | "CANCELED";

export type SocialUser = {
  id: number;
  username: string;
  relationship: RelationshipStatus;
  online: boolean;
};

export type FriendRequest = {
  id: number;
  sender_id: number;
  sender_name: string;
  receiver_id: number;
  receiver_name: string;
  status: FriendRequestStatus;
  created_at: string;
  responded_at?: string;
};

export type Friend = {
  id: number;
  username: string;
  friends_at: string;
  online: boolean;
};

export type FriendNotificationType =
  | "FRIEND_REQUEST_RECEIVED"
  | "FRIEND_REQUEST_ACCEPTED";

export type FriendNotification = {
  id: number;
  type: FriendNotificationType;
  actor_id: number;
  actor_name: string;
  reference_id: number;
  read: boolean;
  read_at?: string;
  created_at: string;
};
