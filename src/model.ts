/* eslint-disable @typescript-eslint/no-explicit-any */

export interface User {
  id: string;
  team_id: string;
  name: string;
  deleted: boolean;
  color: string;
  real_name: string;
  tz: string;
  tz_label: string;
  tz_offset: number;
  profile: any;
  is_admin: boolean;
  is_owner: boolean;
  is_primary_owner: boolean;
  is_restricted: boolean;
  is_ultra_restricted: boolean;
  is_bot: boolean;
  is_app_user: boolean;
  updated: number;
  is_email_confirmed: boolean;
}

export interface Conversation {
  id: string;
  name?: string;
  is_channel?: boolean;
  is_group?: boolean;
  is_im?: boolean;
  created: number;
  is_archived?: boolean;
  is_general?: boolean;
  unlinked?: number;
  name_normalized?: string;
  is_shared?: boolean;
  parent_conversation?: any;
  creator?: string;
  is_ext_shared?: boolean;
  is_org_shared?: boolean;
  shared_team_ids?: any;
  pending_shared?: any[];
  pending_connected_team_ids?: any[];
  is_pending_ext_shared?: boolean;
  is_member?: boolean;
  is_private?: boolean;
  is_mpim?: boolean;
  topic?: any;
  purpose?: any;
  previous_names?: any[];
  num_members?: number;
  user?: string;
  is_user_deleted?: boolean;
  type?: string;
}

export interface SlackMessage {
  client_msg_id: string;
  type: string;
  text: string;
  user: string;
  ts: string;
  team: string;
  blocks: any;
  files?: SlackFile[];
  upload?: boolean;
  display_as_bot?: boolean;
}

// TODO convert into name, text and proper timestamp
export interface SimpleMsg {
  name: string;
  text: string;
  ts: Date;
}

export interface SlackFile {
  id: string;
  created: number;
  timestamp: number;
  name: string;
  title: string;
  mimetype: string;
  filetype: string;
  pretty_type: string;
  user: string;
  editable: boolean;
  size: number;
  mode: string;
  is_external: boolean;
  external_type: string;
  is_public: boolean;
  public_url_shared: boolean;
  display_as_bot: boolean;
  username: string;
  url_private: string;
  url_private_download: string;
  thumb_64: string;
  thumb_80: string;
  thumb_360: string;
  thumb_360_w: number;
  thumb_360_h: number;
  thumb_160: string;
  image_exif_rotation: number;
  original_w: number;
  original_h: number;
  thumb_tiny: string;
  permalink: string;
  permalink_public: string;
  is_starred: boolean;
  has_rich_preview: boolean;
}
