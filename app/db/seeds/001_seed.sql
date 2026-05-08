-- +goose Up
-- +goose StatementBegin

-- ----------------------------------------------------------------
-- users
-- password: password123 (bcrypt DefaultCost)
-- ----------------------------------------------------------------
INSERT INTO users (id, email, password_hash, name, role, department, job_title, status, invited_at, last_login_at, created_at, updated_at)
VALUES
  (1, 'admin@example.com',  '$2a$10$8zNNKZ4yMsihMy88Q20pBefWBH4txT0S/aUv3wJStFllffAE8hC2S', '田中 太郎', 'admin',   '管理部',     'マネージャー',     'active', NOW() - INTERVAL '60 days', NOW() - INTERVAL '1 hour',  NOW() - INTERVAL '60 days', NOW() - INTERVAL '1 hour'),
  (2, 'yamada@example.com', '$2a$10$J2XOp77oaz1CmNxfxU09R.eZ8pZo1zYAVhBEJdF1bdO7/3ILRIlaS', '山田 花子', 'general', '管理部',     '担当者',           'active', NOW() - INTERVAL '55 days', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '55 days', NOW() - INTERVAL '3 hours'),
  (3, 'sato@example.com',   '$2a$10$FOzAAF68CXCh.UjwcRdv5eGVV390/LfV.yQKplE7.p/fh.SlKCS5O', '佐藤 次郎', 'general', '施設管理部', '担当者',           'active', NOW() - INTERVAL '50 days', NOW() - INTERVAL '2 days', NOW() - INTERVAL '50 days', NOW() - INTERVAL '2 days')
ON CONFLICT (id) DO NOTHING;

SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));

-- ----------------------------------------------------------------
-- properties
-- ----------------------------------------------------------------
INSERT INTO properties (id, name, address, area, unit_count, status, management_company, assignee_id, created_by, updated_by, created_at, updated_at)
VALUES
  (1, 'アーバンコート渋谷',         '東京都渋谷区神南1-12-6',         2840.50,  42, 'active',   '山田不動産株式会社',     2, 1, 1, NOW() - INTERVAL '50 days', NOW() - INTERVAL '2 days'),
  (2, 'シティレジデンス新宿',       '東京都新宿区西新宿3-7-1',        5120.00,  68, 'active',   '山田不動産株式会社',     3, 1, 1, NOW() - INTERVAL '45 days', NOW() - INTERVAL '5 days'),
  (3, 'グランドパレス池袋',         '東京都豊島区西池袋2-14-3',       3210.75,  35, 'active',   '東都管理株式会社',       2, 1, 2, NOW() - INTERVAL '40 days', NOW() - INTERVAL '1 day'),
  (4, 'リバーサイドマンション品川', '東京都品川区東品川2-3-12',       4560.25,  55, 'active',   '東都管理株式会社',       3, 1, 3, NOW() - INTERVAL '35 days', NOW() - INTERVAL '3 days'),
  (5, 'ガーデンハイツ目黒',         '東京都目黒区中目黒4-8-2',        2180.00,  28, 'inactive', '山田不動産株式会社', NULL, 1, 1, NOW() - INTERVAL '30 days', NOW() - INTERVAL '10 days')
ON CONFLICT (id) DO NOTHING;

SELECT setval('properties_id_seq', (SELECT MAX(id) FROM properties));

-- ----------------------------------------------------------------
-- claims
-- ----------------------------------------------------------------
INSERT INTO claims (id, title, content, property_id, reporter_name, reporter_contact, status, severity, category, assignee_id, is_recurrence, response_due_at, completed_at, satisfaction_score, created_by, updated_by, created_at, updated_at)
VALUES
  (1,  '給水ポンプから異音がする',
       '昨日から給水ポンプ室付近で金属音が断続的に聞こえます。水圧も若干低下しているように感じます。早急な点検をお願いします。',
       1, '鈴木 一郎', 'suzuki@example.net', 'in_progress', 'high',   '設備',     2, FALSE, NOW() + INTERVAL '2 days',  NULL, NULL, 2, 2, NOW() - INTERVAL '5 days',  NOW() - INTERVAL '1 day'),
  (2,  '共用廊下の照明が切れている',
       '3階の共用廊下、エレベーター前の蛍光灯が1本切れています。夜間は暗くて危険なので早めに交換してください。',
       1, '高橋 恵子', 'takahashi@example.net', 'pending',     'medium', '共用部',   2, FALSE, NOW() + INTERVAL '5 days',  NULL, NULL, NULL, NULL, NOW() - INTERVAL '3 days',  NOW() - INTERVAL '3 days'),
  (3,  '駐輪場の屋根が破損している',
       '先週の台風の影響で駐輪場の屋根の一部が剥がれています。雨天時に自転車が濡れてしまうため、修繕をお願いします。',
       1, '渡辺 健太', 'watanabe@example.net', 'completed',   'medium', '外構',     2, FALSE, NOW() - INTERVAL '10 days', NOW() - INTERVAL '8 days', 4, NULL, 2, NOW() - INTERVAL '15 days', NOW() - INTERVAL '8 days'),
  (4,  'エレベーターの扉が閉まりにくい',
       '2号機のエレベーターの扉が最近閉まるのに時間がかかるようになりました。センサーの不具合でしょうか。点検をお願いします。',
       2, '伊藤 美穂', 'ito@example.net',      'in_progress', 'urgent', '設備',     3, TRUE,  NOW() + INTERVAL '1 day',   NULL, NULL, NULL, 3, NOW() - INTERVAL '2 days',  NOW() - INTERVAL '1 day'),
  (5,  '駐車場区画番号の表示が消えかけている',
       '地下駐車場のP-12からP-15の区画番号のペイントが剥がれてきており、判別が難しい状態です。再塗装をご検討ください。',
       2, '中村 浩二', 'nakamura@example.net', 'pending',     'low',    '外構',  NULL, FALSE, NOW() + INTERVAL '14 days', NULL, NULL, NULL, NULL, NOW() - INTERVAL '1 day',   NOW() - INTERVAL '1 day'),
  (6,  '管理室エアコンの水漏れ',
       '管理室に設置されているエアコンの室内機から水滴が落ちています。ドレンホースの詰まりかと思われます。',
       2, '小林 啓介', 'kobayashi@example.net','completed',   'high',   '設備',     3, FALSE, NOW() - INTERVAL '5 days',  NOW() - INTERVAL '4 days', 5, 3, 3, NOW() - INTERVAL '10 days', NOW() - INTERVAL '4 days'),
  (7,  'ゴミ置き場の扉が壊れた',
       'ゴミ置き場のスライド扉のレールが歪んでいて、開閉が困難な状態です。住民から苦情が来ています。',
       3, '松本 裕子', 'matsumoto@example.net','in_progress', 'medium', '共用部',   2, FALSE, NOW() + INTERVAL '3 days',  NULL, NULL, NULL, 2, NOW() - INTERVAL '4 days',  NOW() - INTERVAL '2 days'),
  (8,  '外壁のひび割れを発見',
       '北側外壁の3〜4階付近に縦方向のひび割れが確認できます。幅は1mm程度ですが、漏水が心配なため調査をお願いします。',
       3, '井上 誠',   'inoue@example.net',    'pending',     'high',   '外壁',  NULL, FALSE, NOW() + INTERVAL '7 days',  NULL, NULL, NULL, NULL, NOW() - INTERVAL '6 days',  NOW() - INTERVAL '6 days'),
  (9,  'オートロックの暗証番号パネルが反応しない',
       'エントランスのオートロックパネルのボタン「5」が押しても反応しないことがあります。暗証番号に5が含まれる住民が困っています。',
       4, '斎藤 良子', 'saito@example.net',    'completed',   'urgent', '設備',     3, FALSE, NOW() - INTERVAL '8 days',  NOW() - INTERVAL '7 days', 3, NULL, 3, NOW() - INTERVAL '12 days', NOW() - INTERVAL '7 days'),
  (10, 'ベランダの排水口が詰まっている',
       '305号室のベランダの排水口が詰まっていて、大雨の際に水が溜まってしまいます。清掃か交換をお願いします。',
       4, '加藤 拓也', 'kato@example.net',     'pending',     'medium', '設備',  NULL, FALSE, NOW() + INTERVAL '10 days', NULL, NULL, NULL, NULL, NOW() - INTERVAL '2 days',  NOW() - INTERVAL '2 days'),
  (11, '自転車の無断駐輪が増えている',
       '共用部の廊下付近に無断で自転車が置かれるケースが増えています。張り紙や警告対応をお願いしたいです。',
       1, '田中 幸子', 'tanaka_s@example.net', 'in_progress', 'low',    'マナー',   2, FALSE, NOW() + INTERVAL '7 days',  NULL, NULL, NULL, 2, NOW() - INTERVAL '7 days',  NOW() - INTERVAL '3 days'),
  (12, '洗濯機置き場の防水パンにひび',
       '208号室の洗濯機置き場の防水パンにひびが入っており、水が染み出す可能性があります。交換をお願いします。',
       2, '木村 達也', 'kimura@example.net',   'pending',     'high',   '設備',  NULL, TRUE,  NOW() + INTERVAL '5 days',  NULL, NULL, NULL, NULL, NOW() - INTERVAL '1 day',   NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

SELECT setval('claims_id_seq', (SELECT MAX(id) FROM claims));

-- ----------------------------------------------------------------
-- claim_responses
-- ----------------------------------------------------------------
INSERT INTO claim_responses (id, claim_id, type, content, created_by, created_at)
VALUES
  (1,  1, 'comment',          '本日、設備業者へ連絡しました。明後日に点検に来てもらえる予定です。', 2, NOW() - INTERVAL '4 days'),
  (2,  1, 'response_history', '業者より「モーターのベアリング摩耗の可能性が高い」との仮診断を受けました。部品手配中です。', 2, NOW() - INTERVAL '2 days'),
  (3,  3, 'comment',          '修繕業者に依頼し、屋根材の交換工事を実施しました。', 2, NOW() - INTERVAL '9 days'),
  (4,  3, 'response_history', '工事完了を確認。居住者へ報告済みです。', 2, NOW() - INTERVAL '8 days'),
  (5,  4, 'comment',          'エレベーター保守会社に緊急連絡しました。明日の午前中に点検予定です。', 3, NOW() - INTERVAL '1 day'),
  (6,  4, 'system_log',       'ステータスを「未対応」から「対応中」に変更しました。', 3, NOW() - INTERVAL '1 day'),
  (7,  6, 'comment',          'ドレンホースの詰まりを解消しました。テスト運転にて水漏れ停止を確認しました。', 3, NOW() - INTERVAL '5 days'),
  (8,  6, 'response_history', '居住者への完了報告を実施しました。満足度評価「5」をいただきました。', 3, NOW() - INTERVAL '4 days'),
  (9,  7, 'comment',          '建具業者に見積もりを依頼しました。今週中に回答が来る予定です。', 2, NOW() - INTERVAL '3 days'),
  (10, 9, 'comment',          'メーカーのサービスマンが訪問し、接点不良を修理しました。', 3, NOW() - INTERVAL '8 days'),
  (11, 9, 'response_history', '動作確認完了。全ボタン正常動作を確認しました。', 3, NOW() - INTERVAL '7 days'),
  (12, 11,'comment',          '共用廊下に「駐輪禁止」の張り紙を掲示し、放置自転車に警告タグを付けました。', 2, NOW() - INTERVAL '5 days')
ON CONFLICT (id) DO NOTHING;

SELECT setval('claim_responses_id_seq', (SELECT MAX(id) FROM claim_responses));

-- ----------------------------------------------------------------
-- internal_memos
-- ----------------------------------------------------------------
INSERT INTO internal_memos (id, claim_id, content, created_by, updated_by, created_at, updated_at)
VALUES
  (1, 1,  '業者への支払いは修繕費予算枠から支出予定。金額は10万円以内に収まる見込み。',                      2, 2, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
  (2, 4,  'これで3回目の同じ不具合。製造から15年経過しており、オーナーへのリニューアル提案を検討する。',      3, 3, NOW() - INTERVAL '1 day',  NOW() - INTERVAL '1 day'),
  (3, 8,  '築22年のため外壁全体の打診調査も合わせて提案したい。来季の大規模修繕計画に盛り込む予定。',        2, 2, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
  (4, 12, '同フロアで過去にも防水パン交換事例あり。経年劣化が進んでいる可能性。他の号室も点検要確認。',      2, 2, NOW() - INTERVAL '1 day',  NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

SELECT setval('internal_memos_id_seq', (SELECT MAX(id) FROM internal_memos));

-- ----------------------------------------------------------------
-- tags
-- ----------------------------------------------------------------
INSERT INTO tags (id, name, created_at)
VALUES
  (1, '緊急対応済み',   NOW() - INTERVAL '30 days'),
  (2, '業者手配中',     NOW() - INTERVAL '25 days'),
  (3, '再発案件',       NOW() - INTERVAL '20 days'),
  (4, '大規模修繕候補', NOW() - INTERVAL '15 days'),
  (5, 'オーナー報告済', NOW() - INTERVAL '10 days')
ON CONFLICT (id) DO NOTHING;

SELECT setval('tags_id_seq', (SELECT MAX(id) FROM tags));

-- ----------------------------------------------------------------
-- claim_tags
-- ----------------------------------------------------------------
INSERT INTO claim_tags (claim_id, tag_id)
VALUES
  (1,  2),
  (4,  1),
  (4,  3),
  (8,  4),
  (9,  1),
  (9,  5),
  (12, 3)
ON CONFLICT DO NOTHING;

-- ----------------------------------------------------------------
-- notifications
-- ----------------------------------------------------------------
INSERT INTO notifications (id, user_id, title, body, link, is_read, created_at)
VALUES
  (1,  1, '新しいクレームが登録されました',         '「洗濯機置き場の防水パンにひび」が登録されました。',                 '/claims/12', FALSE, NOW() - INTERVAL '1 day'),
  (2,  2, '担当クレームのステータスが更新されました','「エレベーターの扉が閉まりにくい」が対応中に更新されました。',       '/claims/4',  FALSE, NOW() - INTERVAL '1 day'),
  (3,  3, '担当クレームのステータスが更新されました','「エレベーターの扉が閉まりにくい」を担当者に割り当てられました。',    '/claims/4',  TRUE,  NOW() - INTERVAL '1 day'),
  (4,  1, 'クレームが完了しました',                 '「管理室エアコンの水漏れ」が完了になりました。満足度: ★★★★★',       '/claims/6',  TRUE,  NOW() - INTERVAL '4 days'),
  (5,  2, '対応期限が近づいています',               '「給水ポンプから異音がする」の対応期限まで2日です。',                 '/claims/1',  FALSE, NOW() - INTERVAL '30 minutes'),
  (6,  1, '新しいクレームが登録されました',         '「ベランダの排水口が詰まっている」が登録されました。',                '/claims/10', TRUE,  NOW() - INTERVAL '2 days'),
  (7,  3, 'クレームが完了しました',                 '「オートロックの暗証番号パネルが反応しない」が完了になりました。',     '/claims/9',  TRUE,  NOW() - INTERVAL '7 days')
ON CONFLICT (id) DO NOTHING;

SELECT setval('notifications_id_seq', (SELECT MAX(id) FROM notifications));

-- ----------------------------------------------------------------
-- audit_logs
-- ----------------------------------------------------------------
INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, before_state, after_state, created_at)
VALUES
  (1,  1, 'create', 'property', 1, NULL,
   '{"name":"アーバンコート渋谷","status":"active"}',
   NOW() - INTERVAL '50 days'),
  (2,  1, 'create', 'property', 2, NULL,
   '{"name":"シティレジデンス新宿","status":"active"}',
   NOW() - INTERVAL '45 days'),
  (3,  2, 'create', 'claim',    1, NULL,
   '{"title":"給水ポンプから異音がする","status":"pending"}',
   NOW() - INTERVAL '5 days'),
  (4,  2, 'update', 'claim',    1,
   '{"status":"pending"}',
   '{"status":"in_progress","assignee_id":2}',
   NOW() - INTERVAL '4 days'),
  (5,  2, 'create', 'claim',    3, NULL,
   '{"title":"駐輪場の屋根が破損している","status":"pending"}',
   NOW() - INTERVAL '15 days'),
  (6,  2, 'update', 'claim',    3,
   '{"status":"in_progress"}',
   '{"status":"completed"}',
   NOW() - INTERVAL '8 days'),
  (7,  3, 'update', 'claim',    4,
   '{"status":"pending"}',
   '{"status":"in_progress","assignee_id":3}',
   NOW() - INTERVAL '1 day'),
  (8,  1, 'update', 'property', 5,
   '{"status":"active"}',
   '{"status":"inactive"}',
   NOW() - INTERVAL '10 days')
ON CONFLICT (id) DO NOTHING;

SELECT setval('audit_logs_id_seq', (SELECT MAX(id) FROM audit_logs));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM audit_logs       WHERE id BETWEEN 1 AND 8;
DELETE FROM notifications    WHERE id BETWEEN 1 AND 7;
DELETE FROM claim_tags       WHERE claim_id IN (1,4,8,9,12);
DELETE FROM tags             WHERE id BETWEEN 1 AND 5;
DELETE FROM internal_memos   WHERE id BETWEEN 1 AND 4;
DELETE FROM claim_responses  WHERE id BETWEEN 1 AND 12;
DELETE FROM claims           WHERE id BETWEEN 1 AND 12;
DELETE FROM properties       WHERE id BETWEEN 1 AND 5;
DELETE FROM users            WHERE id BETWEEN 1 AND 3;
-- +goose StatementEnd
