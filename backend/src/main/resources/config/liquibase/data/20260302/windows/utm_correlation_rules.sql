INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (879, 'Windows: Unusual Process Network Connection', 3, 3, 2, 'Defense Evasion', 'Trusted Developer Utilities Proxy Execution', 'Identifies network activity from unexpected system applications. This may indicate adversarial activity as these applications are often leveraged by adversaries to execute code and evade detection.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1127/"]', 'regexMatch("log.eventDataProcessName", "(Microsoft.Workflow.Compiler.exe|bginfo.exe|cdb.exe|cmstp.exe|csi.exe|dnx.exe|fsi.exe|ieexec.exe|iexpress.exe|odbcconf.exe|rcsi.exe|xwizard.exe)")', '2026-03-02 13:21:40.452980', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataProcessID","operator":"filter_term","value":"{{.log.eventDataProcessID}}"}],"or":null,"within":"now-5m","count":3}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (880, 'Windows: Suspicious Managed Code Hosting Process', 3, 3, 2, 'Defense Evasion', 'T1055 - Process Injection', 'Identifies a suspicious managed code hosting process which could indicate code injection or other form of suspicious code execution.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1055/"]', 'regexMatch("log.eventDataProcessName", "(wscript.exe|cscript.exe|mshta.exe|wmic.exe|regsvr32.exe|svchost.exe|dllhost.exe|cmstp.exe)")', '2026-03-02 13:21:41.849658', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataProcessName","operator":"filter_term","value":"{{.log.eventDataProcessName}}"},{"field":"log.eventdataProcessID","operator":"filter_term","value":"{{.log.eventdataProcessID}}"}],"or":null,"within":"now-5m","count":3}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (881, 'Windows: Potential Credential Access via Renamed COM+ Services DLL', 3, 3, 2, 'Credential Access', 'T1003 - OS Credential Dumping', 'Identifies suspicious renamed COMSVCS.DLL Image Load, which exports the MiniDump function that can be used to dump a process memory. This may indicate an attempt to dump LSASS memory while bypassing command-line based detection in preparation for credential access.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/001/"]', 'contains("log.eventCategory", "process") && contains("log.eventProcessName", "rundll32.exe") && regexMatch("log.eventDataset", "(windows.sysmon_operational)") && equals("log.eventId", "7") && !equals("log.fileName", "COMSVCS.DLL") && (regexMatch("log.filePeOriginalFileName", "(COMSVCS.DLL)") || regexMatch("log.filePeImphash", "(EADBCCBB324829ACB5F2BBE87E5549A8)"))', '2026-03-02 13:21:43.155628', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.processEntityId","operator":"filter_term","value":"{{.log.processEntityId}}"},{"field":"log.eventCategory","operator":"filter_term","value":"process"}],"or":null,"within":"now-5m","count":3}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (882, 'Windows: Multiple Vault Web Credentials Read', 2, 3, 2, 'Credential Access', 'T1555.004 - Credentials from Password Stores: Windows Credential Manager', 'Windows Credential Manager allows you to create, view, or delete saved credentials for signing into websites, connected applications, and networks. An adversary may abuse this to list or dump credentials stored in the Credential Manager for saved usernames and passwords. This may also be performed in preparation of lateral movement.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1555/004/"]', 'equals("log.eventCode", 5382) && !equals("log.eventDataSubjectLogonId", "0x3e7") && (contains("log.eventDataSchemaFriendlyName", "Windows Web Password Credential") || contains("log.eventDataResource", "http"))', '2026-03-02 13:21:44.517513', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataProcessPid","operator":"filter_term","value":"{{.log.eventDataProcessPid}}"},{"field":"log.eventCode","operator":"filter_term","value":"5382"},{"field":"log.eventDataSubjectLogonId","operator":"filter_not_match","value":"0x3e7"}],"or":[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataSchemaFriendlyName","operator":"filter_term","value":"Windows Web Password Credential"},{"field":"log.eventDataResource","operator":"filter_term","value":"http"}],"or":null,"within":"now-60s","count":1}],"within":"now-60s","count":1}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (883, 'Windows: Remote Logon followed by Scheduled Task Creation', 3, 3, 2, 'Lateral Movement', 'T1021 - Remote Services', 'Identifies a remote logon followed by a scheduled task creation on the target host. This could be indicative of adversary lateral movement.', '["https://attack.mitre.org/tactics/TA0008/","https://attack.mitre.org/techniques/T1021/"]', 'equals("log.action", "logged-in") && equals("actionResult", "success") && !contains("log.UserName", "ANONYMOUS LOGON") && !contains("log.UserDomain", "NT AUTHORITY")', '2026-03-02 13:21:45.774870', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.hostId","operator":"filter_term","value":"{{.log.hostId}}"},{"field":"log.eventDataSubjectLogonId","operator":"filter_term","value":"scheduled-task-created"}],"or":null,"within":"now-60s","count":3}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (884, 'Windows: Possible ransomware attack detected. Unusual File Extensions.', 3, 3, 3, 'Impact', 'T1486 - Data Encrypted for Impact', 'Ransomware, is a type of malware that prevents users from accessing their system or  personal files and requires payment of a ransom in order to gain access to them again. Identifies  ransomware attempts. Files with unusual file extensions have been detected, potentially indicating encrypted files created by ransomware.', '["https://attack.mitre.org/tactics/TA0040/"]', 'equals("log.eventCode", 4663) && regexMatch("log.eventDataFileName", "\\.(7z\\.encrypted|aaa|abcd|abtc|acc|aes256|aes_ni|aes256ctr|aes256encrypted|aes_gcm|ajp|alcatraz_locked|alfa|amnesia[1-9]?|amsi|apocalypse666|armadillo|arrow|asasin|asi|atom128|auditor|aurora|autoit|avastvirusinfo|avenger|av666|azero|barak|barrax|bart|beef|beetle|bip|bit|bitcoin|bl2r|blackblock|blackmail|blast|blind|bmw|boot|braincrypt[3-8]?|broken|btcware|budak|bull|buydecryptor|cakl|calipso|calum|carote|cats|cbf|cccmn|ccrypt|cerber[3-9]?|chifrator|chimera|ciphered|crypted|clover|cobra|codnat3|combo|comrade|conficker|coot|cpt|crash|crimson|cry|crypt\\d{3,4}|crypt38|crypt72|crypt888|cryptinfinite|cryptolocker|cryptowall|cryptxxx|crypz|csfr|csone|ctb[1-4]|ctb-locker|ctrsa|cryptowin|cube|dcrpt|ddtf|dr2|dragon|dried|druid|ducky|ecrpt|eeta|etr|ee|f[0-9]{3,4}|flock|grt|grt[0-9]+|grtlock|gwz|h3ll|hades|hakunamatata|hallucinating|happy|harmful|harrow|havoc|headdesk|helpdecrypt|hermes|hidden|hideous|hijack|hilda|hitler|hjg|hmpl|hrosas|hsdf|hushed|hwrm|ihsdj|ikarus|ikasir|ikayed|ill|imbtc|img|encrypted|improved|indrik|injected|innocent|insane|interesante|jungle|kaos|karl|katana|kimcilware|kin|kiratos|kiss|kjh|locked|locky|lokf|losers|lukitus|m3g4c0rtx|m4n1fest0|m4s4g3|maas|madmax|mafi|magic|maktub|malware|manamecrypt|mandelbrot|manic|matrix|max|md5|medusa|mega|melme|merry|mesmerize|metropolitan|mikey|mikibackup|milarepa.lotos@aol.com|mirror|mmnn|mole|monro|mosk|muslat|n1n1n1|nabr|napoleon|narrow|nasoh|nataniel|neitrino|neras|nlah|nosu|novasof|nozelesn|nuclear|nwa|nymaim|obelisk|off|offwhite|ogdo|omega|omerta|onion|ooss|opencode|openme|opqz|osiris|otx|p3rf0rm4|pabluk|pack14|packagetrackr@india.com|packrat|pahd|panda|pandemic|pandora|pansy|paradise|paris|paym3|paymer|payms|pcap|pclock|peet|pelikan|penis|petya|pewcrypt|phoenix|photominr|phobos|phps|pirated|pluto|po1|point|poop|potato|pr0tect|preppy|princesa|princess|prosper|prosperity|prq|pshy|pumas|pumax|pure|purple|purpler|pwnd|pysa|q9q9|qbtex|qiuu|qkkd|qscx|qtyop|quimera|r2d2|r5a|rabit|radman|raid10|rainbow|rakhni|rambo|ramses|rat|rcrypted|react|reactor|realtek|reaper|redlion|redmat|redrum|rekt|remk|removal|remsec|remy|renaming|revenge|rezuc|rhino|ribd|rich|rip|rire|rizonesoft@protonmail.ch|rk|rmdir|robinhood|rocke|rogue|roldat|rolin|ronzware|rosenquist|rotten|roza|rpcminer|rsalive|rumba|run|rxx|uak|udjvu|unlock92|unlckr|upd|urcp|usam|usbc|v8|vag|vandt|varasto|vault|vauw|vb|ve|vendetta|venom|veracrypt|versiegelt|veton|vhd|vindows|violate|virus|vivin|vk_677|vma|vmx|volcano|vorasto|vorphal|vos|vscrypt|vxl|w4b|wakanda|wannacash|wannacry|wanted|war|wasted|wcry|weapologize|webmafia|weird|weui|whatthefuck|whistler|white|whitenoise|whiterabbit|whorus|why!decryptor|wicked|wildfire|windows10|windows7|windows8|windowsupdate|winlock|wipe|wisconsin|wizard|wlu|woolger|worm|wormfubuki|wow|wpencrypt|wq!decryptor|wrui|wtg|x1881|x3m|xampug|xdata|xencrypt|xfiles|xhelper|xlr|xman|xmd|xmd|xtbl|xtbl|xtr|xtt|xtz|xyz|yakuza|yatron|ybn|year|yellow|yheq|yty|yuke|yxo|yyto|z3|zatrov|zax|zbot|zbt|zbt|zeppelin|zerber|zet|zet|zfj|zfj|zimbra|zip|zix|zlz|zobm|zoh|zorro|zphs)$")', '2026-03-02 13:21:47.135287', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventCode","operator":"filter_term","value":"4663"},{"field":"origin.user","operator":"filter_term","value":"{{.origin.user}}"}],"or":null,"within":"now-5m","count":20}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (885, 'Windows: Possible ransomware attack detected. Ransomware Note Creation.', 3, 3, 2, 'Impact', 'T1486 - Data Encrypted for Impact', 'Ransomware, is a type of malware that prevents users from accessing their system or  personal files and requires payment of a ransom in order to gain access to them again. Identifies  ransomware attempts. A known ransomware note file has been detected, potentially indicating an active ransomware infection.', '["https://attack.mitre.org/tactics/TA0040/"]', 'equals("log.eventCode", 4663) && regexMatch("log.EventDataFileName", "(README_TO_RESTORE_FILES|INSTRUCTION_TO_GET_FILES_BACK|HOW_TO_DECRYPT_FILES|DECRYPT_INSTRUCTION|RECOVER_INSTRUCTION|RESTORE_FILES|READ_ME_NOW|YOUR_FILES_ARE_ENCRYPTED|IMPORTANT_INSTRUCTIONS|NOTICE|DECRYPT_YOUR_FILES|HOW_TO_RESTORE_FILES|HELP_DECRYPT|RECOVERY_FILE|RECOVER-FILES|INSTRUCTION)\\.(txt|html|php)$")', '2026-03-02 13:21:48.438769', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventCode","operator":"filter_term","value":"{{.log.eventCode}}"},{"field":"log.EventDataFileName","operator":"filter_term","value":"{{.log.EventDataFileName}}"}],"or":null,"within":"now-60s","count":5}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (886, 'Windows: Possible ransomware attack detected. Multiple File Deletion.', 1, 3, 2, 'Impact', 'T1486 - Data Encrypted for Impact', 'Detects potential ransomware activity by monitoring multiple file write/modification events (Event ID 4663)  with write access masks in user directories within a short timeframe. Modern ransomware typically  encrypts files in-place rather than deleting them, making write access monitoring more effective  than deletion monitoring alone.', '["https://attack.mitre.org/tactics/TA0040/"]', e'equals("log.eventCode", 4663) &&
oneOf("log.eventDataAccessMask", ["0x2", "0x4", "0x6"]) &&
!(regexMatch("log.eventDataProcessName", "(?i).*(trustedinstaller|svchost|wuauclt|msiexec|windows10upgrade|setuphost|tiworker|dism).*")) &&
regexMatch("log.eventDataObjectName", "(?i).*\\\\\\\\(users|documents|desktop|downloads|pictures|videos|music)\\\\\\\\.*") &&
!(regexMatch("log.eventDataObjectName", "(?i).*(\\\\\\\\windows\\\\\\\\|\\\\\\\\program files|\\\\\\\\programdata\\\\\\\\|\\\\\\\\temp\\\\\\\\|\\\\\\\\appdata\\\\\\\\local\\\\\\\\temp|\\\\\\\\softwaredistribution\\\\\\\\|\\\\\\\\winsxs\\\\\\\\|\\\\\\\\logs\\\\\\\\|\\\\\\\\prefetch\\\\\\\\).*")) &&
!(regexMatch("log.eventDataObjectName", "(?i).*\\\\.(tmp|log|etl|dmp|pf|evtx|cache|dat|bak)$")) &&
regexMatch("log.eventDataObjectName", "(?i).*\\\\.(doc[x]?|xls[x]?|ppt[x]?|pdf|txt|jpg|jpeg|png|gif|bmp|mp4|avi|mp3|zip|rar|7z)$")
', '2026-03-02 13:21:49.800468', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventCode","operator":"filter_term","value":"4663"},{"field":"origin.user","operator":"filter_term","value":"{{.origin.user}}"}],"or":null,"within":"now-5m","count":50}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (887, 'Windows: Probable Password guessing', 2, 2, 3, 'Credential Access', 'T1110.001 - Brute Force: Password Guessing', 'Adversaries with no prior knowledge of legitimate credentials within the system or environment  may guess passwords to attempt access to accounts. Without knowledge of the password for an account,  an adversary may opt to systematically guess the password using a repetitive or iterative mechanism.  An adversary may guess login credentials without prior knowledge of system or environment passwords  during an operation by using a list of common passwords. Password guessing may or may not take into  account the target''s policies on password complexity or use policies that may lock accounts out after  a number of failed attempts.', '["https://attack.mitre.org/tactics/TA0006","https://attack.mitre.org/techniques/T1110/001/"]', 'oneOf("log.eventCode", [4625, 529, 530, 531, 532, 533, 534, 535, 536, 537, 539])', '2026-03-02 13:21:51.099967', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventCode","operator":"filter_term","value":"{{.log.eventCode}}"},{"field":"target.user","operator":"filter_term","value":"{{.target.user}}"}],"or":null,"within":"now-5m","count":10}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (888, 'Windows: Multiple TS Gateway login failures', 3, 3, 2, 'Credential Access', 'T1110 - Brute Force', 'Adversaries may use brute force techniques to gain access to accounts when passwords are unknown or when password hashes are obtained.', '["https://attack.mitre.org/techniques/T1110/"]', 'equals("log.eventCode", 1001) && contains("log.message", "Microsoft-Windows-TerminalServices")', '2026-03-02 13:21:52.466461', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-15m","count":10}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (889, 'Windows: Multiple remote access login failures', 3, 2, 2, 'Credential Access', 'T1110 - Brute Force', 'Adversaries may use brute force techniques to gain access to accounts when passwords are unknown or when password hashes are obtained.', '["https://attack.mitre.org/techniques/T1110/"]', e'oneOf("log.eventCode", [20187, 20014, 20078, 20050, 20049, 2018]) &&
regexMatch("log.message", "(?i)(authentication failed|login failed|access denied|authentication error|invalid credentials)")
', '2026-03-02 13:21:53.827075', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-15m","count":10}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (890, 'Windows: Multiple failed attempts to perform a privileged operation by the same user', 1, 2, 3, 'Privilege Escalation', 'T1110 - Brute Force', 'Adversaries may use brute force techniques to gain access to accounts when passwords are unknown or when password hashes are obtained.', '["https://attack.mitre.org/techniques/T1110/"]', 'equals("log.eventCode", 577) || equals("log.eventCode", 4673)', '2026-03-02 13:21:54.875323', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"target.user","operator":"filter_term","value":"{{.target.user}}"}],"or":null,"within":"now-10m","count":10}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (891, 'Windows: Potential Credential Access via Trusted Developer Utility', 2, 2, 3, 'Credential Access', 'T1003 - OS Credential Dumping', 'An instance of MSBuild, the Microsoft Build Engine, loaded DLLs (dynamically linked libraries) responsible for Windows credential management. This technique is sometimes used for credential dumping.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/"]', 'regexMatch("log.eventDataProcessName", "(MSBuild.exe|msbuild.exe)") && regexMatch("log.message", "(vaultcli.dll|SAMLib.DLL)")', '2026-03-02 13:21:56.049626', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataProcessId","operator":"filter_term","value":"{{.log.eventDataProcessId}}"}],"or":null,"within":"now-1m","count":1}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (892, 'Windows: Multiple Logon Failure Followed by Logon Success', 2, 2, 3, 'Credential Access', 'T1110 - Brute Force', 'This rule is triggered when a sequence of multiple failed login attempts followed immediately by a successful login from the same IP address or source is detected. This unusual sequence of events may indicate a possible unauthorized access attempt using a brute force or password guessing technique. The purpose of this rule is to identify suspicious patterns of login activity and alert you to potential unauthorized access attempts.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1110/"]', 'equals("log.eventCode", 4624)', '2026-03-02 13:21:57.308717', true, true, 'origin', '["origin.ip","target.user"]', '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventCode","operator":"filter_term","value":4625},{"field":"target.user","operator":"filter_term","value":"{{.target.user}}"}],"or":null,"within":"now-5m","count":10}]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (893, 'Windows: Possible Brute Force Attack', 2, 2, 3, 'Credential Access', 'T1110 - Brute Force', 'This rule is triggered when a pattern of repeated and rapid login attempts from the same IP address or source is detected. These login attempts may target specific user accounts or services in an attempt to crack passwords through automated brute force. The purpose of this rule is to identify possible malicious unauthorized access attempts and prevent a brute force attack against the system.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1110/"]', 'equals("log.eventCode", 4625)', '2026-03-02 13:21:58.672154', true, true, 'origin', '["origin.host","target.user"]', '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventCode","operator":"filter_term","value":"{{.log.eventCode}}"},{"field":"target.user","operator":"filter_term","value":"{{.target.user}}"}],"or":null,"within":"now-5m","count":10}]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (894, 'Windows: Signed Proxy Execution via MS Work Folders', 1, 2, 3, 'Defense Evasion', 'T1218 - System Binary Proxy Execution', 'Identifies the use of Windows Work Folders to execute a potentially masqueraded control.exe file in the current working directory. Misuse of Windows Work Folders could indicate malicious activity.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1218/"]', 'contains("log.eventDataProcessName", "control.exe") && contains("log.eventDataParentProcessName", "workfolders.exe") && !regexMatch("log.eventDataProcessName", "(:\\Windows\\(System32|SysWOW64)\\control.exe)")', '2026-03-02 13:21:59.853571', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (895, 'Windows: Wireless Credential Dumping using Netsh Command', 3, 3, 2, 'Credential Access', 'T1003 - OS Credential Dumping', 'Identifies attempts to dump Wireless saved access keys in clear text using the Windows built-in utility Netsh.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/"]', 'contains("log.message", "wlan") && regexMatch("log.message", "(key(.+)clear)") && contains("log.eventDataProcessName", "netsh.exe")', '2026-03-02 13:22:01.032912', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (896, 'Windows: Unusual Child Process of dns.exe', 1, 3, 2, 'Initial Access', 'T1133 - External Remote Services', 'Identifies an unexpected process spawning from dns.exe, the process responsible for Windows DNS server services, which may indicate activity related to remote code execution or other forms of exploitation.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1133/"]', 'contains("log.eventDataProcessName", "dns.exe") && !contains("log.eventDataParentProcessName", "conhost.exe")', '2026-03-02 13:22:02.298793', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (897, 'Windows: System Shells via Services', 1, 3, 2, 'Persistence', 'T1543.003 - Create or Modify System Process: Windows Service', 'Windows services typically run as SYSTEM and can be used as a privilege escalation opportunity. Malware or penetration testers may run a shell as a service to gain SYSTEM permissions.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1543/003/"]', e'regexMatch("log.eventDataProcessName", "(cmd.exe|powershell.exe|pwsh.exe|powershell_ise.exe)") &&
 contains("log.eventDataParentProcessName", "services.exe") &&
 !(regexMatch("log.message", "(NVDisplay.ContainerLocalSystem)"))
', '2026-03-02 13:22:03.482260', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (898, 'Windows: Symbolic Link to Shadow Copy Created', 1, 3, 2, 'Credential Access', 'T1003 - OS Credential Dumping', 'Detects creation of a symbolic link to a volume shadow copy. Adversaries may use this technique to access and exfiltrate sensitive data such as NTDS.dit or SAM database from shadow copies.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/"]', 'regexMatch("log.eventDataProcessName", "(cmd.exe|powershell.exe|pwsh.exe|powershell_ise.exe)") && regexMatch("log.message", "(?i)(mklink|New-Item.*SymbolicLink)") && contains("log.message", "HarddiskVolumeShadowCopy")', '2026-03-02 13:22:04.836960', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (899, 'Windows: Suspicious Execution via Scheduled Task', 1, 3, 2, 'Persistence', 'T1053.005 - Scheduled Task', 'Identifies execution of a suspicious program via scheduled tasks by looking at process lineage and command line usage.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1053/005/"]', e'equals("log.eventDataEventType", "start") && contains("log.eventDataProcessParentName", "svchost.exe") && contains("log.eventDataProcessParentArgs", "Schedule") &&
regexMatch("log.eventDataOriginalFileName", "(cscript.exe|wscript.exe|PowerShell.EXE|Cmd.Exe|MSHTA.EXE|RUNDLL32.EXE|REGSVR32.EXE|MSBuild.exe|InstallUtil.exe|RegAsm.exe|RegSvcs.exe|msxsl.exe|CONTROL.EXE|EXPLORER.EXE|Microsoft.Workflow.Compiler.exe|msiexec.exe)") &&
regexMatch("log.eventDataProcessArgs", "(C:\\\\\\\\Users\\\\\\\\|C:\\\\\\\\ProgramData\\\\\\\\|C:\\\\\\\\Windows\\\\\\\\Temp\\\\\\\\|C:\\\\\\\\Windows\\\\\\\\Tasks\\\\\\\\|C:\\\\\\\\PerfLogs\\\\\\\\|C:\\\\\\\\Intel\\\\\\\\|C:\\\\\\\\Windows\\\\\\\\Debug\\\\\\\\|C:\\\\\\\\HP\\\\\\\\)") &&
!regexMatch("log.eventDataProcessName", "(cmd.exe|cscript.exe|powershell.exe|msiexec.exe)") &&
!regexMatch("log.eventDataProcessArgs", "(:\\\\\\\\(.+).bat|:\\\\\\\\Windows\\\\\\\\system32\\\\\\\\calluxxprovider.vbs|-File|-PSConsoleFile)") &&
!regexMatch("log.eventDataUserId", "(S-1-5-18)") && !regexMatch("log.eventDataWorkingDirectory", "(:\\\\\\\\Windows\\\\\\\\System32\\\\\\\\)")
', '2026-03-02 13:22:05.931261', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (900, 'Windows: Suspicious RDP ActiveX Client Loaded', 1, 3, 2, 'Lateral Movement', 'T1021 - Remote Services', 'Identifies suspicious Image Loading of the Remote Desktop Services ActiveX Client (mstscax), this may indicate the presence of RDP lateral movement capability.', '["https://attack.mitre.org/tactics/TA0008/","https://attack.mitre.org/techniques/T1021/"]', e'!(regexMatch("log.eventDataProcessName", "(C:\\\\\\\\Windows\\\\\\\\System32\\\\\\\\mstsc\\\\.exe|C:\\\\\\\\Windows\\\\\\\\SysWOW64\\\\\\\\mstsc\\\\.exe)")) &&
regexMatch("log.eventDataProcessName", "(C:\\\\\\\\Windows\\\\\\\\|C:\\\\\\\\Users\\\\\\\\Public\\\\\\\\|C:\\\\\\\\Users\\\\\\\\Default\\\\\\\\|C:\\\\\\\\Intel\\\\\\\\|C:\\\\\\\\PerfLogs\\\\\\\\|C:\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Device\\\\\\\\Mup\\\\\\\\|\\\\\\\\\\\\\\\\)") &&
contains("log.message", "mstscax.dll")
', '2026-03-02 13:22:07.049460', true, true, 'target', null, '[]', '["target.host","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (901, 'Windows: Suspicious Process Execution via Renamed PsExec Executable', 1, 3, 2, 'Execution', 'T1569 - System Services', 'Identifies suspicious psexec activity which is executing from the psexec service that has been renamed, possibly to vade detection.', '["https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/techniques/T1569/"]', 'equals("log.eventDataEventType", "start") && (contains("log.eventDataProcessName", "PSEXESVC.exe") || contains("log.eventDataOriginalFileName", "psexesvc.exe"))', '2026-03-02 13:22:08.108884', true, true, 'target', null, '[]', '["target.host","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (902, 'Windows: Suspicious Process Access via Direct System Call', 1, 3, 2, 'Defense Evasion', 'T1055 - Process Injection', 'Identifies suspicious process access events from an unknown memory region. Endpoint security solutions usually hook userland Windows APIs in order to decide if the code that is being executed is malicious or not. It''s possible to bypass hooked functions by writing malicious functions that call syscalls directly.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1055/"]', e'equals("log.eventCode", 10) && !(regexMatch("log.eventDataCallTrace", "(:\\\\WINDOWS\\\\SYSTEM32\\\\ntdll.dll|:\\\\WINDOWS\\\\SysWOW64\\\\ntdll.dll|:\\\\Windows\\\\System32\\\\wow64cpu.dll|:\\\\WINDOWS\\\\System32\\\\wow64win.dll|:\\\\Windows\\\\System32\\\\win32u.dll)")) &&
!(regexMatch("log.eventDataTargetImage", "(:\\\\WINDOWS\\\\system32\\\\lsass.exe|:\\\\Program Files (x86)\\\\Malwarebytes Anti-Exploit\\\\mbae-svc.exe|:\\\\Program Files\\\\Cisco\\\\AMP\\\\(.+)\\\\sfc.exe|:\\\\Program Files (x86)\\\\Microsoft\\\\EdgeWebView\\\\Application\\\\(.+)\\\\msedgewebview2.exe|:\\\\Program Files\\\\Adobe\\\\Acrobat DC\\\\Acrobat\\\\(.+)\\\\AcroCEF.exe)")) &&
!(regexMatch("log.eventDataProcessName", "(:\\\\Program Files\\\\Adobe\\\\Acrobat DC\\\\Acrobat\\\\Acrobat.exe|:\\\\Program Files (x86)\\\\World of Warcraft\\\\_classic_\\\\WowClassic.exe)"))
', '2026-03-02 13:22:09.241708', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (903, 'Windows: Suspicious PowerShell Engine ImageLoad', 1, 3, 2, 'Execution', 'T1059 - Command and Scripting Interpreter', 'Identifies the PowerShell engine being invoked by unexpected processes. Rather than executing PowerShell functionality with powershell.exe, some attackers do this to operate more stealthily.', '["https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/techniques/T1059/"]', e'!(oneOf("log.eventDataProcessName", ["Altaro.SubAgent.exe", "AppV_Manage.exe", "azureadconnect.exe", "CcmExec.exe", "configsyncrun.exe", "choco.exe", "ctxappvservice.exe", "DVLS.Console.exe", "edgetransport.exe", "exsetup.exe", "forefrontactivedirectoryconnector.exe", "InstallUtil.exe", "JenkinsOnDesktop.exe", "Microsoft.EnterpriseManagement.ServiceManager.UI.Console.exe", "mmc.exe", "mscorsvw.exe", "msexchangedelivery.exe", "msexchangefrontendtransport.exe", "msexchangehmworker.exe", "msexchangesubmission.exe", "msiexec.exe", "MsiExec.exe", "noderunner.exe", "NServiceBus.Host.exe", "NServiceBus.Host32.exe", "NServiceBus.Hosting.Azure.HostProcess.exe", "OuiGui.WPF.exe", "powershell.exe", "powershell_ise.exe", "pwsh.exe", "SCCMCliCtrWPF.exe", "ScriptEditor.exe", "ScriptRunner.exe", "sdiagnhost.exe", "servermanager.exe", "setup100.exe", "ServiceHub.VSDetouredHost.exe", "SPCAF.Client.exe", "SPCAF.SettingsEditor.exe", "SQLPS.exe", "telemetryservice.exe", "UMWokerProcess.exe", "w3wp.exe", "wsmprovhost.exe"])) &&
!(regexMatch("log.eventDataProcessName", "(C:\\\\Windows\\\\System32\\\\RemoteFXvGPUDisablement.exe|C:\\\\Windows\\\\System32\\\\sdiagnhost.exe|C:\\\\Program Files( \\\\(x86\\\\))?\\\\(.+)\\\\.exe)")) &&
oneOf("log.message", ["System.Management.Automation.ni.dll", "System.Management.Automation.dll"])
', '2026-03-02 13:22:10.641905', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (904, 'Windows: Suspicious PDF Reader Child Process', 1, 1, 2, 'Execution', 'T1204 - User Execution', 'Identifies suspicious child processes of PDF reader applications. These child processes are often launched via exploitation of PDF applications or social engineering.', '["https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/techniques/T1204/"]', 'oneOf("log.eventDataParentProcessName", ["AcroRd32.exe", "Acrobat.exe", "FoxitPhantomPDF.exe", "FoxitReader.exe"]) && oneOf("log.eventDataProcessName", ["arp.exe", "dsquery.exe", "dsget.exe", "gpresult.exe", "hostname.exe", "ipconfig.exe", "nbtstat.exe", "net.exe", "net1.exe", "netsh.exe", "netstat.exe", "nltest.exe", "ping.exe", "qprocess.exe", "quser.exe", "qwinsta.exe", "reg.exe", "sc.exe", "systeminfo.exe", "tasklist.exe", "tracert.exe", "whoami.exe", "bginfo.exe", "cdb.exe", "cmstp.exe", "csi.exe", "dnx.exe", "fsi.exe", "ieexec.exe", "iexpress.exe", "installutil.exe", "Microsoft.Workflow.Compiler.exe", "msbuild.exe", "mshta.exe", "msxsl.exe", "odbcconf.exe", "rcsi.exe", "regsvr32.exe", "xwizard.exe", "atbroker.exe", "forfiles.exe", "schtasks.exe", "regasm.exe", "regsvcs.exe", "cmd.exe", "cscript.exe", "powershell.exe", "pwsh.exe", "wmic.exe", "wscript.exe", "bitsadmin.exe", "certutil.exe", "ftp.exe"])', '2026-03-02 13:22:11.904183', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (905, 'Windows: Suspicious MS Outlook Child Process', 1, 3, 2, 'Initial Access', 'T1566.001 - Phishing: Spearphishing Attachment', 'Identifies suspicious child processes of Microsoft Outlook. These child processes are often associated with spear phishing activity.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1566/001/"]', 'oneOf("log.eventDataProcessName", ["Microsoft.Workflow.Compiler.exe", "arp.exe", "atbroker.exe", "bginfo.exe", "bitsadmin.exe", "cdb.exe", "certutil.exe", "cmd.exe", "cmstp.exe", "cscript.exe", "csi.exe", "dnx.exe", "dsget.exe", "dsquery.exe", "forfiles.exe", "fsi.exe", "ftp.exe", "gpresult.exe", "hostname.exe", "ieexec.exe", "iexpress.exe", "installutil.exe", "ipconfig.exe", "mshta.exe", "msxsl.exe", "nbtstat.exe", "net.exe", "net1.exe", "netsh.exe", "netstat.exe", "nltest.exe", "odbcconf.exe", "ping.exe", "powershell.exe", "pwsh.exe", "qprocess.exe", "quser.exe", "qwinsta.exe", "rcsi.exe", "reg.exe", "regasm.exe", "regsvcs.exe", "regsvr32.exe", "sc.exe", "schtasks.exe", "systeminfo.exe", "tasklist.exe", "tracert.exe", "whoami.exe", "wmic.exe", "wscript.exe", "xwizard.exe"]) && contains("log.eventDataParentProcessName", "outlook.exe")', '2026-03-02 13:22:13.266068', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (906, 'Windows: Suspicious MS Office Child Process', 1, 2, 3, 'Initial Access', 'T1566.001 - Phishing: Spearphishing Attachment', 'Identifies suspicious child processes of frequently targeted Microsoft Office applications (Word, PowerPoint, Excel). These child processes are often launched during exploitation of Office applications or from documents with malicious macros.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1566/001/"]', 'oneOf("log.eventDataProcessName", ["Microsoft.Workflow.Compiler.exe", "arp.exe", "atbroker.exe", "bginfo.exe", "bitsadmin.exe", "cdb.exe", "certutil.exe", "cmd.exe", "cmstp.exe", "control.exe", "cscript.exe", "csi.exe", "dnx.exe", "dsget.exe", "dsquery.exe", "forfiles.exe", "fsi.exe", "ftp.exe", "gpresult.exe", "hostname.exe", "ieexec.exe", "iexpress.exe", "installutil.exe", "ipconfig.exe", "mshta.exe", "msxsl.exe", "nbtstat.exe", "net.exe", "net1.exe", "netsh.exe", "netstat.exe", "nltest.exe", "odbcconf.exe", "ping.exe", "powershell.exe", "pwsh.exe", "qprocess.exe", "quser.exe", "qwinsta.exe", "rcsi.exe", "reg.exe", "regasm.exe", "regsvcs.exe", "regsvr32.exe", "sc.exe", "schtasks.exe", "systeminfo.exe", "tasklist.exe", "tracert.exe", "whoami.exe", "wmic.exe", "wscript.exe", "xwizard.exe", "explorer.exe", "rundll32.exe", "hh.exe", "msdt.exe"]) && oneOf("log.eventDataParentProcessName", ["eqnedt32.exe", "excel.exe", "fltldr.exe", "msaccess.exe", "mspub.exe", "powerpnt.exe", "winword.exe", "outlook.exe"])', '2026-03-02 13:22:14.575639', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (907, 'Windows: Microsoft Exchange Worker Spawning Suspicious Processes', 2, 3, 2, 'Initial Access', 'T1190 - Exploit Public-Facing Application', 'Identifies suspicious processes being spawned by the Microsoft Exchange Server worker process (w3wp). This activity may indicate exploitation activity or access to an existing web shell backdoor.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1190/"]', 'contains("log.eventDataParentProcessName", "w3wp.exe") && regexMatch("log.message", "(MSExchange(.+)AppPool)") && regexMatch("log.eventDataProcessName", "(md.exe|powershell.exe|pwsh.dll|powershell_ise.exe)")', '2026-03-02 13:22:15.888276', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (908, 'Windows: Microsoft Exchange Worker Spawning Suspicious Processes', 2, 3, 2, 'Initial Access', 'T1190 - Exploit Public-Facing Application', 'Identifies suspicious processes being spawned by the Microsoft Exchange Server worker process (w3wp). This activity may indicate exploitation activity or access to an existing web shell backdoor.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1190/"]', 'regexMatch("log.eventDataParentProcessName", "(UMService.exe|UMWorkerProcess.exe)") && !(regexMatch("log.eventDataProcessName", "(:\\Windows\\System32\\werfault.exe|:\\Windows\\System32\\wermgr.exe|:\\Program Files\\Microsoft\\Exchange Server\\V(.+)\\Bin\\UMWorkerProcess.exe|D:\\Exchange 2016\\Bin\\UMWorkerProcess.exe|E:\\ExchangeServer\\Bin\\UMWorkerProcess.exe)"))', '2026-03-02 13:22:17.113022', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (909, 'Windows: Potential LSASS Memory Dump via PssCaptureSnapShot', 3, 3, 2, 'Credential Access', 'T1003 - OS Credential Dumping', 'Identifies suspicious access to an LSASS handle via PssCaptureSnapShot where two successive process accesses are performed by the same process and target two different instances of LSASS. This may indicate an attempt to evade detection and dump LSASS memory for credential access.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/"]', 'equals("log.eventId", "10") && regexMatch("log.eventDataTargetImage", "([Cc]:\\Windows\\[Ss]ystem32\\lsass.exe)")', '2026-03-02 13:22:18.376805', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (910, 'Windows: Potential Credential Access via LSASS Memory Dump', 3, 3, 2, 'Credential Access', 'T1003 - OS Credential Dumping', 'Identifies suspicious access to LSASS handle from a call trace pointing to DBGHelp.dll or DBGCore.dll, which both export the MiniDumpWriteDump method that can be used to dump LSASS memory content in preparation for credential access.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/"]', 'equals("log.eventId", "10") && regexMatch("log.eventDataTargetImage", "(C:\\\\WINDOWS\\\\system32\\\\lsass.exe)") && regexMatch("log.eventDataCallTrace", "(dbghelp|dbgcore)") && !regexMatch("log.eventDataProcessName", "(C:\\\\Windows\\\\System32\\\\WerFault.exe|C:\\\\Windows\\\\System32\\\\WerFaultSecure.exe)")', '2026-03-02 13:22:19.775611', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (911, 'Windows: Remote Computer Account DnsHostName Update', 3, 3, 2, 'Privilege Escalation', 'T1068 - Exploitation for Privilege Escalation', 'Identifies the remote update to a computer account''s DnsHostName attribute. If the new value set is a valid domain controller DNS hostname and the subject computer name is not a domain controller, then it''s highly likely a preparation step to exploit CVE-2022-26923 in an attempt to elevate privileges from a standard domain user to domain admin privileges.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1068/"]', 'equals("log.action", "logged-in") && regexMatch("actionResult", "success") && !contains("log.userName", "ANONYMOUS LOGON") && !contains("log.eventDataSubjectUserName", "ANONYMOUS LOGON") && startsWith("log.eventDataSubjectUserName", "$") && !contains("log.userDomain", "NT AUTHORITY")', '2026-03-02 13:22:21.133002', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (912, 'Windows: Service Control Spawned via Script Interpreter', 2, 3, 2, 'Lateral Movement', 'T1021 - Remote Services', 'Identifies Service Control (sc.exe) spawning from script interpreter processes to create, modify, or start services. This could be indicative of adversary lateral movement but will be noisy if commonly done by admins.', '["https://attack.mitre.org/tactics/TA0008/","https://attack.mitre.org/techniques/T1021/"]', 'oneOf("log.eventDataParentProcessName", ["cmd.exe", "wscript.exe", "rundll32.exe", "regsvr32.exe", "wmic.exe", "mshta.exe", "powershell.exe", "pwsh.exe"]) && oneOf("log.message", ["config", "create", "start", "delete", "stop", "pause"]) && !equals("log.eventDataSubjectUserName", "S-1-5-18") && contains("log.eventDataProcessName", "sc.exe")', '2026-03-02 13:22:22.441377', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (913, 'Windows: Sensitive Privilege SeEnableDelegationPrivilege assigned to a User', 3, 3, 1, 'Credential Access', 'T1212 - Exploitation for Credential Access', 'Identifies the assignment of the SeEnableDelegationPrivilege sensitive user right to a user. The SeEnableDelegationPrivilege user right enables computer and user accounts to be trusted for delegation. Attackers can abuse this right to compromise Active Directory accounts and elevate their privileges.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1212/"]', 'regexMatch("action", "([Aa]uthorization [Pp]olicy [Cc]hange)") && equals("log.eventCode", 4704) && contains("log.eventDataPrivilegeList", "SeEnableDelegationPrivilege")', '2026-03-02 13:22:23.793649', true, true, 'target', null, '[]', '["target.host","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (914, 'Windows: Security Software Discovery using WMIC', 3, 2, 1, 'Discovery', 'T1518.001 - Software Discovery: Security Software Discovery', 'Identifies the use of Windows Management Instrumentation Command (WMIC) to discover certain System Security Settings such as AntiVirus or Host Firewall details.', '["https://attack.mitre.org/tactics/TA0007/","https://attack.mitre.org/techniques/T1518/001/"]', 'regexMatch("log.message", "(namespace:\\\\root\\SecurityCenter2)") && contains("log.message", "Get") && contains("log.eventDataProcessName", "wmic.exe")', '2026-03-02 13:22:25.092283', true, true, 'origin', '["adversary.user","adversary.ip"]', '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (915, 'Windows: Potential Secure File Deletion via SDelete Utility', 3, 2, 2, 'Defense Evasion', 'T1070.004 - Indicator Removal: File Deletion', 'Detects file name patterns generated by the use of Sysinternals SDelete utility to securely delete a file via multiple file overwrite and rename operations.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1070/004/"]', 'equals("log.eventDataEventType", "change") && contains("log.eventDataFileName", "AAA.AAA")', '2026-03-02 13:22:26.331342', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (916, 'Windows Script Interpreter Executing Process via WMI', 2, 2, 2, 'Initial Access', 'T1566.001 - Phishing: Spearphishing Attachment', 'Identifies use of the built-in Windows script interpreters (cscript.exe or wscript.exe) being used to execute a process via Windows Management Instrumentation (WMI). This may be indicative of malicious activity.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1566/001/"]', 'oneOf("log.processName", ["wscript.exe", "cscript.exe"]) && contains("log.message", "wmiutils.dll")', '2026-03-02 13:22:27.599095', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (917, 'Windows Script Executing PowerShell', 2, 3, 2, 'Initial Access', 'T1566.001 - Phishing: Spearphishing Attachment', 'Identifies a PowerShell process launched by either cscript.exe or wscript.exe. Observing Windows scripting processes executing a PowerShell script, may be indicative of malicious activity.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1566/001/"]', 'equals("log.eventDataEventType", "start") && oneOf("log.eventDataParentProcessName", ["cscript.exe", "wscript.exe"]) && contains("log.eventDataProcessName", "powershell.exe")', '2026-03-02 13:22:28.773925', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (918, 'Windows: Remote Scheduled Task Creation', 2, 2, 2, 'Lateral Movement', 'T1021 - Remote Services', 'Identifies remote scheduled task creations on a target host. This could be indicative of adversary lateral movement.', '["https://attack.mitre.org/tactics/TA0008/","https://attack.mitre.org/techniques/T1021/"]', 'contains("log.processName", "svchost.exe") && oneOf("log.eventDataSourceNetworkAddress", ["incoming", "ingress"]) && greaterOrEqual("origin.port", 49152) && greaterOrEqual("target.port", 49152) && !oneOf("origin.ip", ["127.0.0.1", "::1"])', '2026-03-02 13:22:30.128739', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (919, 'Windows: Outbound Scheduled Task Activity via PowerShell', 2, 3, 2, 'Execution', 'T1053.005 - Scheduled Task', 'Identifies the PowerShell process loading the Task Scheduler COM DLL followed by an outbound RPC network connection within a short time period. This may indicate lateral movement or remote discovery via scheduled tasks.', '["https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/techniques/T1053/005/"]', 'oneOf("log.eventDataProcessName", ["powershell.exe", "pwsh.exe", "powershell_ise.exe"]) && regexMatch("log.message", "(powershell.exe|pwsh.exe|powershell_ise.exe)")', '2026-03-02 13:22:31.388488', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (920, 'Windows: Searching for Saved Credentials via VaultCmd', 3, 2, 1, 'Credential Access', 'T1555 - Credentials from Password Stores', 'Windows Credential Manager allows you to create, view, or delete saved credentials for signing into websites, connected applications, and networks. An adversary may abuse this to list or dump credentials stored in the Credential Manager for saved usernames and passwords. This may also be performed in preparation of lateral movement.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1555/"]', 'contains("log.eventDataProcessName", "vaultcmd.exe") && contains("log.message", "/list")', '2026-03-02 13:22:32.623128', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (921, 'Windows: Potential Privileged Escalation via SamAccountName Spoofing', 2, 3, 1, 'Privilege Escalation', 'T1078 - Valid Accounts', 'Identifies a suspicious computer account name rename event, which may indicate an attempt to exploit CVE-2021-42278 to elevate privileges from a standard domain user to a user with domain admin privileges. CVE-2021-42278 is a security vulnerability that allows potential attackers to impersonate a domain controller via samAccountName attribute spoofing.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1078/"]', 'equals("action", "renamed-user-account") && endsWith("target.user", "$") && !endsWith("log.eventDataNewTargetUserName", "$")', '2026-03-02 13:22:33.799924', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (922, 'Windows: Execution of Persistent Suspicious Program', 2, 3, 2, 'Persistence', 'T1547 - Boot or Logon Autostart Execution', 'Identifies execution of suspicious persistent programs (scripts, rundll32, etc.) by looking at process lineage and command line usage', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1547/"]', 'contains("log.eventDataProcessName", "explorer.exe")', '2026-03-02 13:22:35.066139', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (923, 'Windows: Unusual Child Processes of RunDLL32', 2, 3, 2, 'Defense Evasion', 'T1218.011 - System Binary Proxy Execution: Rundll32', 'Identifies child processes of unusual instances of RunDLL32 where the command line parameters were suspicious. Misuse of RunDLL32 could indicate malicious activity', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1218/011/"]', 'contains("log.eventDataProcessName", "rundll32.exe")', '2026-03-02 13:22:36.151604', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (924, 'Windows: Remote System Discovery Commands', 3, 1, 2, 'Discovery', 'T1018 - Remote System Discovery', 'Discovery of remote system information using built-in commands, which may be used to move laterally.', '["https://attack.mitre.org/tactics/TA0007/","https://attack.mitre.org/techniques/T1018/"]', 'regexMatch("log.message", "(-n|-s|-a)") && regexMatch("log.eventDataProcessName", "(nbtstat.exe|arp.exe)")', '2026-03-02 13:22:37.340444', true, true, 'origin', '["adversary.user","adversary.ip"]', '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (925, 'Windows: Remote File Download via MpCmdRun', 2, 3, 1, 'Command and Control', 'T1105 - Ingress Tool Transfer', 'Identifies the Windows Defender configuration utility (MpCmdRun.exe) being used to download a remote file.', '["https://attack.mitre.org/tactics/TA0011/","https://attack.mitre.org/techniques/T1105/"]', 'contains("log.eventDataProcessName", "MpCmdRun.exe") && contains("log.message", "-url") && contains("log.message", "-DownloadFile") && contains("log.message", "-path")', '2026-03-02 13:22:38.562990', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (926, 'Windows: Remote File Copy to a Hidden Share', 3, 2, 1, 'Lateral Movement', 'T1021 - Remote Services', 'Identifies a remote file copy attempt to a hidden network share. This may indicate lateral movement or data staging activity.', '["https://attack.mitre.org/tactics/TA0008/","https://attack.mitre.org/techniques/T1021/"]', 'oneOf("log.eventDataProcessName", ["cmd.exe", "powershell.exe", "robocopy.exe", "xcopy.exe"]) && contains("log.message", "$") && (contains("log.message", "copy") || contains("log.message", "move") || contains("log.message", " cp ") || contains("log.message", " mv "))', '2026-03-02 13:22:39.745107', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (927, 'Windows: Remote File Download via Desktopimgdownldr Utility', 2, 3, 1, 'Command and Control', 'T1105 - Ingress Tool Transfer', 'Identifies the desktopimgdownldr utility being used to download a remote file. An adversary may use desktopimgdownldr to download arbitrary files as an alternative to certutil.', '["https://attack.mitre.org/tactics/TA0011/","https://attack.mitre.org/techniques/T1105/"]', 'contains("log.eventDataProcessName", "desktopimgdownldr.exe")', '2026-03-02 13:22:41.055446', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (928, 'Windows: Suspicious Microsoft Diagnostics Wizard Execution with args IT_RebrowseForFile=, ms-msdt:/id, ms-msdt:-id or FromBase64', 2, 3, 1, 'Defense Evasion', 'T1218 - System Binary Proxy Execution', 'Identifies potential abuse of the Microsoft Diagnostics Troubleshooting Wizard (MSDT) to proxy malicious command or binary execution via malicious process arguments', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1218/"]', 'contains("log.eventDataProcessName", "msdt.exe") && regexMatch("log.message", "(IT_RebrowseForFile=|ms-msdt:/id|ms-msdt:-id|FromBase64)")', '2026-03-02 13:22:42.316588', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (929, 'Windows: Privilege Escalation via Rogue Named Pipe Impersonation', 1, 3, 2, 'Privilege Escalation', 'T1134 - Access Token Manipulation', 'Identifies a privilege escalation attempt via rogue named pipe impersonation. An adversary may abuse this technique by masquerading as a known named pipe and manipulating a privileged process to connect to it.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1134/"]', 'regexMatch("log.eventDataProcessName", "(\\(.+)\\Pipe\\)") && contains("log.action", "Pipe Created")', '2026-03-02 13:22:43.676556', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (930, 'Windows: Potential Modification of Accessibility Binaries', 2, 3, 1, 'Persistence', 'T1546.008 - Event Triggered Execution: Accessibility Features', 'Windows contains accessibility features that may be launched with a key combination before a user has logged in. An adversary can modify the way these programs are launched to get a command prompt or backdoor without logging in to the system.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1546/008/"]', e'!regexMatch("log.eventDataProcessName", "(osk.exe|sethc.exe|utilman2.exe|DisplaySwitch.exe|ATBroker.exe|ScreenMagnifier.exe|SR.exe|Narrator.exe|magnify.exe|MAGNIFY.EXE)") &&
  regexMatch("log.eventDataParentProcessName", "(Utilman.exe|on.exe)") &&
  contains("log.eventDataSubjectUserName", "SYSTEM") &&
  regexMatch("log.message", "(C:\\\\Windows\\\\System32\\\\osk.exe|C:\\\\Windows\\\\System32\\\\Magnify.exe|C:\\\\Windows\\\\System32\\\\Narrator.exe|C:\\\\Windows\\\\System32\\\\Sethc.exe|utilman.exe|ATBroker.exe|DisplaySwitch.exe|sethc.exe)")
', '2026-03-02 13:22:44.854436', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (931, 'Windows: Suspicious Print Spooler SPL File Created', 1, 3, 2, 'Privilege Escalation', 'T1068 - Exploitation for Privilege Escalation', 'Detects attempts to exploit privilege escalation vulnerabilities related to the Print Spooler service including CVE-2020-1048 and CVE-2020-1337.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1068/"]', e'!regexMatch("log.eventDataProcessName", "(spoolsv.exe|printfilterpipelinesvc.exe|PrintIsolationHost.exe|splwow64.exe|msiexec.exe|poqexec.exe)") && regexMatch("log.eventDataProcessName", "(:\\\\Windows\\\\System32\\\\spool\\\\PRINTERS\\\\)")
', '2026-03-02 13:22:46.156806', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (932, 'Windows: Suspicious PrintSpooler Service Executable File Creation', 2, 3, 1, 'Privilege Escalation', 'T1068 - Exploitation for Privilege Escalation', 'Detects attempts to exploit privilege escalation vulnerabilities related to the Print Spooler service. For more information refer to the following CVE''s - CVE-2020-1048, CVE-2020-1337 and CVE-2020-1300 and verify that the impacted system is patched', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1068/"]', e'!regexMatch("log.file.path", "(\\\\Windows\\\\System32\\\\spool\\\\|:\\\\Windows\\\\Temp\\\\|:\\\\Users\\\\)") && contains("log.eventDataProcessName", "spoolsv.exe")
', '2026-03-02 13:22:47.519189', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (933, 'Windows Firewall Disabled via PowerShell', 1, 2, 3, 'Defense Evasion', 'T1562.004 - Impair Defenses: Disable or Modify System Firewall', 'Identifies when the Windows Firewall is disabled using PowerShell cmdlets, which can help attackers evade network constraints, like internet and network lateral communication restrictions.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/004/"]', 'regexMatch("log.message", "(-All|Public|Domain|Private)") && contains("log.message", "False") && contains("log.message", "-Enabled") && contains("log.message", "Set-NetFirewallProfile") && regexMatch("log.eventDataProcessName", "(powershell.exe|pwsh.exe|powershell_ise.exe)")', '2026-03-02 13:22:48.749324', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (934, 'Windows: User logged using Remote Desktop Connection from loopback address, possible exploit over reverse tunneling using stolen credentials', 3, 2, 1, 'Credential Access', 'T1021.001 - Remote Services: Remote Desktop Protocol', 'Adversaries may use Valid Accounts to log into a computer using the Remote Desktop Protocol (RDP). The adversary may then perform actions as the logged-on user.', '["https://attack.mitre.org/techniques/T1021/001/"]', 'equals("log.eventDataLogonType", "10") && oneOf("log.origin.ips", ["::1", "127.0.0.1"]) && oneOf("log.eventCode", [528, 540, 673, 4624, 4769])', '2026-03-02 13:22:50.137480', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (935, 'Windows: Persistence via WMI Event Subscription', 2, 3, 2, 'Persistence', 'T1546.003 - Event Triggered Execution: WMI Event Subscription', 'An adversary can use Windows Management Instrumentation (WMI) to install event filters, providers, consumers, and bindings that execute code when a defined event occurs. Adversaries may use the capabilities of WMI to subscribe to an event and execute arbitrary code when that event occurs, providing persistence on a system.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1546/003/"]', 'contains("log.message", "create") && regexMatch("log.message", "(ActiveScriptEventConsumer|CommandLineEventConsumer)") && contains("log.eventDataProcessName", "wmic.exe")', '2026-03-02 13:22:51.268994', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (936, 'Windows: Persistence via Update Orchestrator Service Hijack', 2, 3, 2, 'Persistence', 'T1543.003 - Create or Modify System Process: Windows Service', 'Identifies potential hijacking of the Microsoft Update Orchestrator Service to establish persistence with an integrity level of SYSTEM.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1543/003/"]', e'contains("log.message", "UsoSvc") &&
 regexMatch("log.eventDataParentProcessName", "C:\\\\\\\\Windows\\\\\\\\System32\\\\\\\\svchost\\\\.exe") &&
 !regexMatch("log.eventDataProcessName", "MoUsoCoreWorker\\\\.exe|OfficeC2RClient\\\\.exe") &&
 !regexMatch("log.eventDataProcessName", "ProgramData\\\\\\\\Microsoft\\\\\\\\Windows\\\\\\\\UUS\\\\\\\\Packages\\\\\\\\.*\\\\\\\\amd64\\\\\\\\MoUsoCoreWorker\\\\.exe|Windows\\\\\\\\System32\\\\\\\\UsoClient\\\\.exe|Windows\\\\\\\\System32\\\\\\\\MusNotification\\\\.exe|Windows\\\\\\\\System32\\\\\\\\MusNotificationUx\\\\.exe|Windows\\\\\\\\System32\\\\\\\\MusNotifyIcon\\\\.exe|Windows\\\\\\\\System32\\\\\\\\WerFault\\\\.exe|Windows\\\\\\\\System32\\\\\\\\WerMgr\\\\.exe|Windows\\\\\\\\UUS\\\\\\\\amd64\\\\\\\\MoUsoCoreWorker\\\\.exe|Windows\\\\\\\\System32\\\\\\\\MoUsoCoreWorker\\\\.exe|Windows\\\\\\\\UUS\\\\\\\\amd64\\\\\\\\UsoCoreWorker\\\\.exe|Windows\\\\\\\\System32\\\\\\\\UsoCoreWorker\\\\.exe|Program Files\\\\\\\\Common Files\\\\\\\\microsoft shared\\\\\\\\ClickToRun\\\\\\\\OfficeC2RClient\\\\.exe")
', '2026-03-02 13:22:52.636684', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (937, 'Windows: Persistence via TelemetryController Scheduled Task Hijack', 2, 3, 2, 'Persistence', 'T1053.005 - Scheduled Task', 'Detects the successful hijack of Microsoft Compatibility Appraiser scheduled task to establish persistence with an integrity level of system.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1053/005/"]', 'contains("log.message", "-cv") && contains("log.eventDataParentProcessName", "CompatTelRunner.exe") && !(regexMatch("log.eventDataProcessName", "(conhost.exe|DeviceCensus.exe|CompatTelRunner.exe|DismHost.exe|rundll32.exe|powershell.exe)"))', '2026-03-02 13:22:53.941140', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (938, 'Windows: Persistence via BITS Job Notify Cmdline', 2, 3, 2, 'Persistence', 'T1197 - BITS Jobs', 'An adversary can use the Background Intelligent Transfer Service (BITS) SetNotifyCmdLine method to execute a program that runs after a job finishes transferring data or after a job enters a specified state in order to persist on a system.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1197/"]', 'contains("log.message", "BITS") && contains("log.eventDataParentProcessName", "svchost.exe") && !(regexMatch("log.eventDataProcessName", "(:\\Windows\\System32\\WerFaultSecure.exe|:\\Windows\\System32\\WerFault.exe|:\\Windows\\System32\\wermgr.exe|:\\WINDOWS\\system32\\directxdatabaseupdater.exe)"))', '2026-03-02 13:22:55.295776', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (939, 'Windows: Privilege Escalation via Named Pipe Impersonation', 3, 3, 2, 'Privilege Escalation', 'T1134 - Access Token Manipulation', 'Identifies a privilege escalation attempt via named pipe impersonation. An adversary may abuse this technique by utilizing a framework such Metasploit''s meterpreter getsystem command.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1134/"]', 'regexMatch("log.eventDataProcessName", "(Cmd.Exe|PowerShell.EXE|powershell.exe|cmd.exe)") && contains("log.message", ">") && regexMatch("log.message", "(\\\\.\\pipe\\)")', '2026-03-02 13:22:56.549601', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (940, 'Windows: Mounting Hidden or WebDav Remote Shares', 3, 2, 1, 'Lateral Movement', 'T1021.002 - Remote Services: SMB/Windows Admin Shares', 'Identifies the use of net.exe to mount a WebDav or hidden remote share. This may indicate lateral movement or preparation for data exfiltration.', '["https://attack.mitre.org/tactics/TA0008/","https://attack.mitre.org/techniques/T1021/002/"]', 'regexMatch("log.eventDataProcessName", "(net.exe|net1.exe)") && regexMatch("log.message", "(\\\\(.+)\\(.+)$|\\\\(.+)@SSL\\|http)") && contains("log.message", "/d") && contains("log.message", "use") && !(contains("log.eventDataParentProcessName", "net.exe"))', '2026-03-02 13:22:57.640797', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (941, 'Windows: Modification of Boot Configuration', 1, 3, 3, 'Impact', 'T1490 - Inhibit System Recovery', 'Identifies use of bcdedit.exe to delete boot configuration data. This tactic is sometimes used as by malware or an attacker as a destructive technique.', '["https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1490/"]', 'contains("log.eventDataProcessName", "bcdedit.exe") && regexMatch("log.message", "((ignoreallfailures(.+)bootstatuspolicy(.+)/set)|(ignoreallfailures(.+)/set(.+)bootstatuspolicy)|(/set(.+)bootstatuspolicy(.+)ignoreallfailures)|(/set(.+)ignoreallfailures(.+)bootstatuspolicy)|(bootstatuspolicy(.+)set(.+)ignoreallfailures)|(bootstatuspolicy(.+)ignoreallfailures(.+)/set)|(no(.+)recoveryenabled)|(recoveryenabled(.+)no))")', '2026-03-02 13:22:58.966800', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (942, 'Windows: Microsoft security essentials - Virus detected', 3, 3, 2, 'Privilege Escalation', 'T1055 - Process Injection', 'Detect the presence of a virus or malware on the system using Microsoft Security Essentials.  The rule correlates different threat detection events, represented by various Event IDs, to identify virus detection on the system.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1055/"]', 'oneOf("log.eventCode", [1107, 1117, 1116, 1118, 1119]) && equals("log.providerName", "Microsoft Antimalware")', '2026-03-02 13:23:00.266586', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (943, 'Windows: LSASS Memory Dump Handle Access', 2, 3, 3, 'Credential Access', 'T1003.001 - OS Credential Dumping: LSASS Memory', 'Identifies handle requests for the Local Security Authority Subsystem Service (LSASS) object access with specific access masks that many tools with a capability to dump memory to disk use (0x1fffff, 0x1010, 0x120089). This rule is tool agnostic as it has been validated against a host of various LSASS dump tools such as SharpDump, Procdump, Mimikatz, Comsvcs etc. It detects this behavior at a low level and does not depend on a specific tool or dump file name.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/","https://attack.mitre.org/techniques/T1003/001/"]', 'equals("log.eventCode", 4656) && regexMatch("log.eventDataObjectName", "(:\\Windows\\System32\\lsass.exe|\\Device\\HarddiskVolume[A-Za-z?:\\]([A-Za-z?])?\\Windows\\System32\\lsass.exe)") && !regexMatch("log.eventDataProcessName", "(:\\Program Files\\(.+).exe|:\\Program Files (x86)\\(.+).exe|:\\Windows\\system32\\wbem\\WmiPrvSE.exe|:\\Windows\\System32\\dllhost.exe|:\\Windows\\System32\\svchost.exe|:\\Windows\\System32\\msiexec.exe|:\\ProgramData\\Microsoft\\Windows Defender\\(.+).exe|:\\Windows\\explorer.exe)") && oneOf("log.eventDataAccessMask", ["0x1fffff", "0x1010", "0x120089", "0x1F3FFF"])', '2026-03-02 13:23:01.542431', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (944, 'Windows: IIS HTTP Logging Disabled', 3, 2, 3, 'Defense Evasion', 'T1562.002 - Impair Defenses: Disable Windows Event Logging', 'Identifies when Internet Information Services (IIS) HTTP Logging is disabled on a server. An attacker with IIS server access via a webshell or other mechanism can disable HTTP Logging as an effective anti-forensics measure.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/002/"]', 'contains("log.eventDataProcessName", "appcmd.exe") && regexMatch("log.message", "(/dontLog(.+):(.+)True)") && !(contains("log.eventDataParentProcessName", "iissetup.exe"))', '2026-03-02 13:23:02.680765', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (945, 'Windows: Microsoft IIS Connection Strings Decryption', 2, 3, 1, 'Credential Access', 'T1003 - OS Credential Dumping', 'Identifies use of aspnet_regiis to decrypt Microsoft IIS connection strings. An attacker with Microsoft IIS web server access via a webshell or alike can decrypt and dump any hardcoded connection strings, such as the MSSQL service account password using aspnet_regiis command.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/"]', 'regexMatch("log.eventDataProcessName", "aspnet_regiis.exe") && regexMatch("log.message", "(connectionStrings)") && regexMatch("log.message", "(-pdf)")', '2026-03-02 13:23:03.856518', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (946, 'Windows: Microsoft IIS Service Account Password Dumped', 3, 2, 2, 'Credential Access', 'T1003 - OS Credential Dumping', 'Identifies the Internet Information Services (IIS) command-line tool, AppCmd, being used to list passwords. An attacker with IIS web server access via a web shell can decrypt and dump the IIS AppPool service account password using AppCmd.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/"]', 'regexMatch("log.eventDataProcessName", "appcmd.exe") && regexMatch("log.message", "(/list)") && regexMatch("log.message", "(/text(.+)password)")', '2026-03-02 13:23:05.170767', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (947, 'Windows: Possible sideloading of DLL via Microsoft antimalware service executable with MsMpEng process', 3, 3, 3, 'Defense Evasion', 'T1574 - Hijack Execution Flow', 'Identifies a Windows trusted program that is known to be vulnerable to DLL Search Order Hijacking starting after being renamed or from a non-standard path. This is uncommon behavior and may indicate an attempt to evade defenses via side-loading a malicious DLL within the memory space of one of those processes.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1574/"]', 'contains("log.eventDataProcessName", "MsMpEng.exe") && !regexMatch("log.eventDataProcessPath", "(:\\ProgramData\\Microsoft\\Windows Defender\\(.+).exe|:\\Program Files\\Windows Defender\\(.+).exe|:\\Program Files (x86)\\Windows Defender\\(.+).exe|:\\Program Files\\Microsoft Security Client\\(.+).exe|:\\Program Files (x86)\\Microsoft Security Client\\(.+).exe)")', '2026-03-02 13:23:06.299379', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (948, 'Windows: Execution via MSSQL xp_cmdshell Stored Procedure', 3, 3, 2, 'Execution', 'T1059 - Command and Scripting Interpreter', 'Identifies execution via MSSQL xp_cmdshell stored procedure. Malicious users may attempt to elevate their privileges by using xp_cmdshell, which is disabled by default, thus, it''s important to review the context of it''s use.', '["https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/techniques/T1059/"]', 'oneOf("log.message", ["diskfree", "rmdir", "mkdir", "dir", "del", "rename", "bcp", "XMLNAMESPACES"]) && contains("log.eventDataProcessName", "cmd.exe") && contains("log.eventDataParentProcessName", "sqlservr.exe")', '2026-03-02 13:23:07.566968', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (949, 'Windows: Process Activity via Compiled HTML File', 3, 3, 1, 'Execution', 'T1204.002 - User Execution: Malicious File', 'Compiled HTML files (.chm) are commonly distributed as part of the Microsoft HTML Help system. Adversaries may conceal malicious code in a CHM file and deliver it to a victim for execution. CHM content is loaded by the HTML Help executable program (hh.exe).', '["https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/techniques/T1204/002/"]', 'regexMatch("log.eventDataProcessName", "(mshta.exe|cmd.exe|powershell.exe|pwsh.exe|powershell_ise.exe|cscript.exe|wscript.exe)") && regexMatch("log.eventDataParentProcessName", "hh.exe")', '2026-03-02 13:23:08.842441', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (950, 'Windows: Microsoft Build Engine Started an Unusual Process', 3, 3, 2, 'Defense Evasion', 'T1027 - Obfuscated Files or Information', 'An instance of MSBuild, the Microsoft Build Engine, started a PowerShell script or the Visual C# Command Line Compiler. This technique is sometimes used to deploy a malicious payload using the Build Engine.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1027/"]', 'regexMatch("log.eventDataProcessName", "(csc.exe|iexplore.exe|powershell.exe)") && regexMatch("log.eventDataParentProcessName", "MSBuild.exe")', '2026-03-02 13:23:10.028196', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (951, 'Windows: Microsoft Build Engine Started by a System Process', 3, 3, 2, 'Defense Evasion', 'T1127.001 - Trusted Developer Utilities Proxy Execution: MSBuild', 'An instance of MSBuild, the Microsoft Build Engine, was started by Explorer or the WMI (Windows Management Instrumentation) subsystem. This behavior is unusual and is sometimes used by malicious payloads.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1127/001/"]', 'regexMatch("log.eventDataProcessName", "MSBuild.exe") && regexMatch("log.eventDataParentProcessName", "(explorer.exe|wmiprvse.exe)")', '2026-03-02 13:23:11.157181', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (952, 'Windows: Microsoft Build Engine Started by a Script Process', 3, 3, 2, 'Defense Evasion', 'T1127.001 - Trusted Developer Utilities Proxy Execution: MSBuild', 'An instance of MSBuild, the Microsoft Build Engine, was started by a script or the Windows command interpreter. This behavior is unusual and is sometimes used by malicious payloads.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1127/001/"]', 'regexMatch("log.eventDataProcessName", "MSBuild.exe") && regexMatch("log.eventDataParentProcessName", "(cmd.exe|powershell.exe|wscript.exe|cscript.exe)")', '2026-03-02 13:23:12.468332', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (953, 'Windows: Microsoft Build Engine Started by an Office Application', 3, 3, 1, 'Defense Evasion', 'T1127.001 - Trusted Developer Utilities Proxy Execution: MSBuild', 'An instance of MSBuild, the Microsoft Build Engine, was started by Excel or Word. This is unusual behavior for the Build Engine and could have been caused by an Excel or Word document executing a malicious script payload.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1127/001/"]', 'regexMatch("log.eventDataProcessName", "MSBuild.exe") && regexMatch("log.eventDataParentProcessName", "(eqnedt32.exe|excel.exe|fltldr.exe|msaccess.exe|mspub.exe|outlook.exe|powerpnt.exe|winword.exe)")', '2026-03-02 13:23:13.606773', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (954, 'Windows: Remote Desktop Enabled in Windows Firewall by Netsh', 3, 3, 1, 'Defense Evasion', 'T1562.004 - Impair Defenses: Disable or Modify System Firewall', 'Identifies use of the network shell utility (netsh.exe) to enable inbound Remote Desktop Protocol (RDP) connections in the Windows Firewall.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/004/"]', 'regexMatch("log.eventDataProcessName", "netsh.exe") && regexMatch("log.message", "(action=allow|enable=Yes|enable)") && regexMatch("log.message", "(localport=3389|RemoteDesktop|group=(.+)remote desktop(.+))")', '2026-03-02 13:23:14.870837', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (955, 'Windows: Detection of Exchange mail exported through PowerShell', 3, 1, 1, 'Collection', 'T1005 - Data from Local System', 'This rule identifies the use of an Exchange PowerShell cmdlet, which is used to export the contents of a core file. mailbox or archive to a .pst file. Adversaries can target user email to collect sensitive information.', '["https://attack.mitre.org/tactics/TA0009/","https://attack.mitre.org/techniques/T1005/","https://attack.mitre.org/techniques/T1114/","https://attack.mitre.org/techniques/T1114/002/"]', 'regexMatch("log.eventDataProcessName", "(powershell.exe|pwsh.exe|powershell_ise.exe)") && regexMatch("log.message", "(New-MailboxExportRequest)")', '2026-03-02 13:23:16.225231', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (956, 'Windows: Credential Acquisition via Registry Hive Dumping', 3, 1, 1, 'Credential Access', 'T1003.002 - OS Credential Dumping: Security Account Manager', 'Identifies attempts to export a registry hive which may contain credentials using the Windows reg.exe tool.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/","https://attack.mitre.org/techniques/T1003/002/"]', 'regexMatch("log.eventDataProcessName", "reg.exe") && regexMatch("log.message", "(save|export)") && regexMatch("log.message", "(hklm\\sam|hklm\\security)")', '2026-03-02 13:23:17.528615', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (957, 'Windows: Disabling Windows Defender Security Settings via PowerShell', 3, 3, 3, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', 'Identifies use of the Set-MpPreference PowerShell command to disable or weaken certain Windows Defender settings.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/001/"]', 'regexMatch("log.message", "(Set-MpPreference)") && regexMatch("log.message", "(-Disable|Disabled|NeverSend|-Exclusion)") && regexMatch("log.eventDataProcessName", "(powershell.exe|pwsh.dll|powershell_ise.exe)")', '2026-03-02 13:23:18.845833', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (958, 'Windows: Disable Windows Firewall Rules via Netsh', 2, 2, 3, 'Defense Evasion', 'T1562.004 - Impair Defenses: Disable or Modify System Firewall', 'Identifies use of netsh.exe to disable Windows Firewall rules or turn off the firewall entirely. Adversaries may disable the Windows Firewall to enable network connections for lateral movement or command and control.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/004/"]', 'regexMatch("log.message", "(disable(.+)firewall(.+)set|disable(.+)set(.+)firewall|firewall(.+)disable(.+)set|firewall(.+)set(.+)disable|set(.+)disable(.+)firewall|set(.+)firewall(.+)disable|state(.+)advfirewall(.+)off|state(.+)off(.+)advfirewall|advfirewall(.+)state(.+)off|advfirewall(.+)off(.+)state|off(.+)state(.+)advfirewall|off(.+)advfirewall(.+)state)") && regexMatch("log.eventDataProcessName", "netsh.exe")', '2026-03-02 13:23:20.152852', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (959, 'Windows: PowerShell Keylogging Script', 3, 2, 2, 'Collection', 'T1056.001 - Input Capture: Keylogging', 'Detects the use of Win32 API Functions that can be used to capture user keystrokes in PowerShell scripts. Attackers use this technique to capture user input, looking for credentials and/or other valuable data.', '["https://attack.mitre.org/tactics/TA0009/","https://attack.mitre.org/techniques/T1056/","https://attack.mitre.org/techniques/T1056/001/"]', 'regexMatch("log.message", "(GetAsyncKeyState|NtUserGetAsyncKeyState|GetKeyboardState|Get-Keystrokes|SetWindowsHookA|SetWindowsHookW|SetWindowsHookEx|SetWindowsHookExA|NtUserSetWindowsHookEx|GetForegroundWindow|GetWindowTextA|GetWindowTextW|WM_KEYBOARD_LL)") && regexMatch("log.eventDataProcessName", "(powershell.exe|pwsh.exe|powershell_ise.exe)")', '2026-03-02 13:23:21.466419', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (960, 'Windows: Deleting Backup Catalogs with Wbadmin', 1, 2, 3, 'Impact', 'T1490 - Inhibit System Recovery', 'Identifies use of the wbadmin.exe to delete the backup catalog. Ransomware and other malware may do this to prevent system recovery.', '["https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1490/"]', 'regexMatch("log.message", "(delete(.+)catalog|catalog(.+)delete)") && regexMatch("log.eventDataProcessName", "(wbadmin.exe|WBADMIN.EXE)")', '2026-03-02 13:23:22.729718', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (961, 'Windows: Delete Volume USN Journal with Fsutil', 1, 2, 3, 'Defense Evasion', 'T1070.004 - Indicator Removal: File Deletion', 'Identifies use of the fsutil.exe to delete the volume USNJRNL. This technique is used by attackers to eliminate evidence of files created during post-exploitation activities.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1070/004/"]', 'regexMatch("log.message", "(deletejournal(.+)usn|usn(.+)deletejournal)") && regexMatch("log.eventDataProcessName", "fsutil.exe")', '2026-03-02 13:23:24.045793', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (962, 'Windows Defender Exclusions Added via PowerShell', 2, 2, 3, 'Defense Evasion', 'T1562 - Impair Defenses', 'Identifies modifications to the Windows Defender configuration settings using PowerShell to add exclusions at the folder directory or process level.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/"]', 'regexMatch("log.message", "(-Exclusion(.+)(Add-MpPreference|Set-MpPreference)|(Add-MpPreference|Set-MpPreference)(.+)-Exclusion)") && regexMatch("log.eventDataProcessName", "(powershell.exe|pwsh.exe|powershell_ise.exe)")', '2026-03-02 13:23:25.350856', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (963, 'Windows: NTDS or SAM Database File Copied', 3, 3, 3, 'Credential Access', 'T1003.002 - OS Credential Dumping: Security Account Manager', 'Identifies a copy operation of the Active Directory Domain Database or Security Account Manager (SAM) files. Those files contain sensitive information including hashed domain and local credentials.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/","https://attack.mitre.org/techniques/T1003/002/"]', 'regexMatch("log.eventDataProcessName", "(?i)(cmd\\.exe|powershell\\.exe|xcopy\\.exe|esentutl\\.exe)") && regexMatch("log.message", "(copy|xcopy|Copy-Item|move|cp|mv|/y|/vss|/d)") && regexMatch("log.message", "(\\ntds.dit|\\config\\SAM|\\(.+)\\GLOBALROOT\\Device\\HarddiskVolumeShadowCopy(.+)\\|/system32/config/SAM)")', '2026-03-02 13:23:26.709174', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (964, 'Clearing Windows Event Logs with wevtutil', 1, 2, 3, 'Defense Evasion', 'T1070.001 - Indicator Removal: Clear Windows Event Logs', 'Identifies attempts to clear or disable Windows event log stores using Windows wevetutil command. This is often done by attackers in an attempt to evade detection or destroy forensic evidence on a system.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1070/001/"]', 'regexMatch("log.message", "(/e:false|cl|clear-log|Clear-EventLog)") && regexMatch("log.eventDataLogonProcessName", "wevtutil.exe")', '2026-03-02 13:23:27.940941', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (965, 'Windows: Clearing Windows Console History', 1, 2, 3, 'Defense Evasion', 'T1070.003 - Indicator Removal: Clear Command History', 'Identifies when a user attempts to clear console history. An adversary may clear the command history of a compromised account to conceal the actions undertaken during an intrusion.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1070/003/"]', 'regexMatch("log.message", "(Clear-History|(Remove-Item|rm)(.+)(ConsoleHost_history.txt|\\(Get-PSReadlineOption\\)\\.HistorySavePath)|(ConsoleHost_history.txt|\\(Get-PSReadlineOption\\)\\.HistorySavePath)(.+)(Remove-Item|rm)|Set-PSReadlineOption(.+)SaveNothing|SaveNothing(.+)PSReadlineOption)") && equals("log.providerName", "PowerShell")', '2026-03-02 13:23:29.287886', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (966, 'Windows Defender: Antimalware engine found malware or other potentially unwanted software', 1, 2, 3, 'Execution', 'T1546 - Event Triggered Execution', 'This rule is triggered when the antimalware engine detects malware or potentially unwanted software on the system. This alert is critical to identify the presence of threats and unwanted software that may compromise system security and performance.', '["https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/techniques/T1546/"]', 'oneOf("log.eventCode", [1006, 1015, 1116]) && equals("log.providerName", "SecurityCenter")', '2026-03-02 13:23:30.591494', true, true, 'target', null, '[]', '["target.host","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (967, 'Windows Service Installed via an Unusual Client', 3, 3, 2, 'Privilege Escalation', 'T1543.003 - Create or Modify System Process: Windows Service', 'Identifies the creation of a Windows service by an unusual client process. Services may be created with administrator privileges but are executed under SYSTEM privileges, so an adversary may also use a service to escalate privileges from administrator to SYSTEM.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1543/003/"]', 'equals("action", "service-installed") && equals("clientProcessId", "0") && equals("parentProcessId", "0")', '2026-03-02 13:23:31.951538', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (968, 'Windows: Whoami Process Activity', 1, 1, 0, 'Discovery', 'T1033 - System Owner/User Discovery', 'Identifies suspicious use of whoami.exe which displays user, group, and privileges information for the user who is currently logged on to the local system.', '["https://attack.mitre.org/tactics/TA0007/","https://attack.mitre.org/techniques/T1033/"]', 'contains("log.eventDataProcessName", "whoami.exe")', '2026-03-02 13:23:33.260121', true, true, 'origin', '["adversary.user","adversary.ip"]', '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (969, 'Windows Web Shell Detection: Script Process Child of Common Web Processes', 1, 3, 2, 'Persistence', 'T1505.003 - Server Software Component: Web Shell', 'Identifies suspicious commands executed via a web server, which may suggest a vulnerability and remote shell access.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1505/003/"]', 'regexMatch("log.eventDataProcessName", "(cmd.exe|cscript.exe|powershell.exe|pwsh.exe|powershell_ise.exe|wmic.exe|wscript.exe)") && regexMatch("log.eventDataParentProcessName", "(w3wp.exe|httpd.exe|nginx.exe|php.exe|php-cgi.exe|tomcat.exe)")', '2026-03-02 13:23:34.618192', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (970, 'Windows: Volume Shadow Copy Deletion via WMIC', 1, 2, 3, 'Impact', 'T1490 - Inhibit System Recovery', 'Identifies use of wmic.exe for shadow copy deletion on endpoints. This commonly occurs in tandem with ransomware or other destructive attacks.', '["https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1490/"]', 'regexMatch("log.message", "(delete(.+)shadowcopy|shadowcopy(.+)delete)") && contains("log.eventDataProcessName", "WMIC.exe")', '2026-03-02 13:23:35.668424', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (971, 'Windows: Volume Shadow Copy Deletion via PowerShell', 1, 2, 3, 'Impact', 'T1490 - Inhibit System Recovery', 'Identifies the use of the Win32_ShadowCopy class and related cmdlets to achieve shadow copy deletion. This commonly occurs in tandem with ransomware or other destructive attacks.', '["https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1490/"]', 'regexMatch("log.eventDataProcessName", "(powershell.exe|pwsh.exe|powershell_ise.exe)") && regexMatch("log.message", "(Get-WmiObject|gwmi|Get-CimInstance|gcim)") && regexMatch("log.message", "(Win32_ShadowCopy)") && regexMatch("log.message", "(\\.Delete\\(\\)|Remove-WmiObject|rwmi|Remove-CimInstance|rcim)")', '2026-03-02 13:23:36.718635', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (972, 'Windows: Volume Shadow Copy Deleted or Resized via VssAdmin', 1, 2, 3, 'Impact', 'T1490 - Inhibit System Recovery', 'Identifies use of vssadmin.exe for shadow copy deletion or resizing on endpoints. This commonly occurs in tandem with ransomware or other destructive attacks.', '["https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1490/"]', 'regexMatch("log.message", "((delete|resize)(.+)shadows|shadows(.+)(delete|resize))") && contains("log.eventDataProcessName", "vssadmin.exe")', '2026-03-02 13:23:37.827087', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (973, 'Windows: Unusual Service Host Child Process - Childless Service', 1, 3, 2, 'Privilege Escalation', 'T1055 - Process Injection', 'Identifies unusual child processes of Service Host (svchost.exe) that traditionally do not spawn any child processes. This may indicate a code injection or an equivalent form of exploitation.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1055/"]', 'contains("log.eventDataParentProcessName", "svchost.exe") && regexMatch("log.message", "(WdiSystemHost|LicenseManager|StorSvc|CDPSvc|cdbhsvc|BthAvctpSvc|SstpSvc|WdiServiceHost|imgsvc|TrkWks|WpnService|IKEEXT|PolicyAgent|CryptSvc|netprofm|ProfSvc|StateRepository|camsvc|LanmanWorkstation|NlaSvc|EventLog|hidserv|DisplayEnhancementService|ShellHWDetection|AppHostSvc|fhsvc|CscService|PushToInstall)") && !regexMatch("log.eventDataProcessName", "(WerFault.exe|WerFaultSecure.exe|wermgr.exe|rundll32.exe)") && !regexMatch("log.eventDataProcessName", "(:\\Windows\\System32\\RelPost.exe|:\\Program Files\\|:\\Program Files (x86)\\|:\\Windows\\System32\\Kodak\\kds_i4x50\\lib\\lexexe.exe)") && !regexMatch("log.message", "(WdiSystemHost|WdiServiceHost|imgsvc)") && !regexMatch("log.message", "(:\\WINDOWS\\System32\\winethc.dll,ForceProxyDetectionOnNextRun)")', '2026-03-02 13:23:38.900423', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (974, 'Windows: Unusual Print Spooler Child Process', 1, 3, 2, 'Privilege Escalation', 'T1068 - Exploitation for Privilege Escalation', 'Detects unusual Print Spooler service (spoolsv.exe) child processes. This may indicate an attempt to exploit privilege escalation vulnerabilities related to the Printing Service on Windows.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1068/"]', 'contains("log.eventDataParentProcessName", "spoolsv.exe") && !regexMatch("log.eventDataProcessName", "(splwow64.exe|PDFCreator.exe|acrodist.exe|spoolsv.exe|msiexec.exe|route.exe|WerFault.exe|net.exe|cmd.exe|powershell.exe|netsh.exe|regsvr32.exe)") && !regexMatch("log.message", "(\\WINDOWS\\system32\\spool\\DRIVERS|stop|start|.spl|\\program files(.+)route add|add portopening|rule name|PrintConfig.dll)") && equals("log.logName", "System")', '2026-03-02 13:23:40.035451', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (975, 'Windows: Unusual Network Connection via DllHost or via RunDLL32', 2, 3, 2, 'Defense Evasion', 'T1218 - System Binary Proxy Execution', 'Identifies unusual instances of dllhost.exe making outbound network connections. This may indicate adversarial Command and Control activity. Identifies unusual instances of rundll32.exe making outbound network connections. This may indicate adversarial Command and Control activity.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1218/"]', 'regexMatch("log.eventDataProcessName", "(dllhost.exe|rundll32.exe)") && regexMatch("log.origin.ips", "((10.0.0.0/8,127.0.0.0/8,169.254.0.0/16,172.16.0.0/12,192.0.0.0/24,192.0.0.0/29,192.0.0.8/32,192.0.0.9/32,192.0.0.10/32,192.0.0.170/32,192.0.0.171/32,192.0.2.0/24,192.31.196.0/24,192.52.193.0/24,192.168.0.0/16,192.88.99.0/24,224.0.0.0/4,100.64.0.0/10,192.175.48.0/24,198.18.0.0/15,198.51.100.0/24,203.0.113.0/24,240.0.0.0/4,::1,FE80::/10,FF00::/8)")', '2026-03-02 13:23:41.348436', true, true, 'target', null, '[]', '["target.host","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (976, 'Windows: Unusual File Modification by dns.exe', 1, 3, 2, 'Initial Access', 'T1133 - External Remote Services', 'Identifies an unexpected file being modified by dns.exe, the process responsible for Windows DNS Server services, which may indicate activity related to remote code execution or other forms of exploitation.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1133/"]', 'contains("log.eventDataProcessName", "dns.exe") && !contains("log.eventDataFileName", "dns.log")', '2026-03-02 13:23:42.439086', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (977, 'Windows: Unusual File Creation - Alternate Data Stream', 1, 3, 2, 'Defense Evasion', 'T1564 - Hide Artifacts', 'Identifies suspicious creation of Alternate Data Streams on highly targeted files. This is uncommon for legitimate files and sometimes done by adversaries to hide malware.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1564/"]', 'regexMatch("log.eventDataFileName", "(^C:\\(.+):(.+))") && !regexMatch("log.eventDataFileName", "(C:\\(.+):zone.identifier)") && regexMatch("log.message", "(pdf|dll|png|exe|dat|com|bat|cmd|sys|vbs|ps1|hta|txt|vbe|js|wsh|docx|doc|xlsx|xls|pptx|ppt|rtf|gif|jpg|png|bmp|img|iso)") && !regexMatch("log.eventDataProcessName", "(:\\windows\\System32\\svchost.exe|:\\Windows\\System32\\inetsrv\\w3wp.exe|:\\Windows\\explorer.exe|:\\Windows\\System32\\sihost.exe|:\\Windows\\System32\\PickerHost.exe|:\\Windows\\System32\\SearchProtocolHost.exe|:\\Program Files (x86)\\Dropbox\\Client\\Dropbox.exe|:\\Program Files\\Rivet Networks\\SmartByte\\SmartByteNetworkService.exe|:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe|:\\Program Files\\ExpressConnect\\ExpressConnectNetworkService.exe|:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe|:\\Program Files\\Google\\Chrome\\Application\\chrome.exe|:\\Program Files\\Mozilla Firefox\\firefox.exe)")', '2026-03-02 13:23:43.618034', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (978, 'Windows: UAC Bypass via Windows Firewall Snap-In Hijack', 1, 3, 2, 'Privilege Escalation', 'T1548.002 - Abuse Elevation Control Mechanism: Bypass User Account Control', 'Identifies attempts to bypass User Account Control (UAC) by hijacking the Microsoft Management Console (MMC) Windows Firewall snap-in. Attackers bypass UAC to stealthily execute code with elevated permissions.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1548/002/"]', 'contains("log.eventDataProcessName", "mmc.exe") && !contains("log.message", "WerFault.exe") && contains("log.message", "WF.msc")', '2026-03-02 13:23:44.791393', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (979, 'Windows: Bypass UAC via Event Viewer', 1, 3, 2, 'Privilege Escalation', 'T1548.002 - Abuse Elevation Control Mechanism: Bypass User Account Control', 'Identifies User Account Control (UAC) bypass via eventvwr.exe. Attackers bypass UAC to stealthily execute code with elevated permissions.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1548/002/"]', 'contains("log.eventDataParentProcessName", "eventvwr.exe") && !regexMatch("log.eventDataProcessName", "(:\\Windows\\SysWOW64\\mmc.exe|:\\Windows\\System32\\mmc.exe|:\\Windows\\SysWOW64\\WerFault.exe|:\\Windows\\System32\\WerFault.exe)")', '2026-03-02 13:23:46.023108', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (980, 'Windows: UAC Bypass Attempt via Privileged IFileOperation COM Interface', 1, 3, 2, 'Privilege Escalation', 'T1548.002 - Abuse Elevation Control Mechanism: Bypass User Account Control', 'Identifies attempts to bypass User Account Control (UAC) via DLL side-loading. Attackers may attempt to bypass UAC to stealthily execute code with elevated permissions.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1548/002/"]', 'regexMatch("log.eventDataFileName", "(wow64log.dll|comctl32.dll|DismCore.dll|OskSupport.dll|duser.dll|Accessibility.ni.dll)") && contains("log.eventDataProcessName", "dllhost.exe") && !regexMatch("log.eventDataFileName", "(C:\\Windows\\SoftwareDistribution\\|C:\\Windows\\WinSxS\\)")', '2026-03-02 13:23:47.252476', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (981, 'Windows: UAC Bypass via ICMLuaUtil Elevated COM Interface', 1, 3, 2, 'Privilege Escalation', 'T1548.002 - Abuse Elevation Control Mechanism: Bypass User Account Control', 'Identifies User Account Control (UAC) bypass attempts via the ICMLuaUtil Elevated COM interface. Attackers may attempt to bypass UAC to stealthily execute code with elevated permissions', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1548/002/"]', 'contains("log.eventDataProcessName", "dllhost.exe") && !contains("log.message", "WerFault.exe") && regexMatch("log.message", "(/Processid:\\{3E5FC7F9-9A51-4367-9063-A120244FBEC7\\}|/Processid:\\{D2E7041B-2927-42FB-8E9F-7CE93B6DC937\\})")', '2026-03-02 13:23:48.644376', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (982, 'Windows: UAC Bypass Attempt via Elevated COM Internet Explorer Add-On Installer', 1, 3, 2, 'Privilege Escalation', 'T1548.002 - Abuse Elevation Control Mechanism: Bypass User Account Control', 'Identifies User Account Control (UAC) bypass attempts by abusing an elevated COM Interface to launch a malicious program. Attackers may attempt to bypass UAC to stealthily execute code with elevated permissions.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1548/002/"]', 'regexMatch("log.message", "(C:\\(.+)\\AppData\\(.+)\\Temp\\IDC(.+).tmp\\(.+).exe)") && contains("log.processParentName", "ieinstall.exe") && regexMatch("log.message", "(-Embedding)")', '2026-03-02 13:23:49.992699', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (983, 'Windows: UAC Bypass Attempt with IEditionUpgradeManager Elevated COM Interface', 1, 3, 2, 'Privilege Escalation', 'T1548.002 - Abuse Elevation Control Mechanism: Bypass User Account Control', 'Identifies attempts to bypass User Account Control (UAC) by abusing an elevated COM Interface to launch a rogue Windows ClipUp program. Attackers may attempt to bypass UAC to stealthily execute code with elevated permissions.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1548/002/"]', 'contains("log.eventDataProcessName", "Clipup.exe") && !regexMatch("log.message", "(C:\\Windows\\System32\\ClipUp.exe)") && contains("log.eventDataParentProcessName", "dllhost.exe") && regexMatch("log.message", "(/Processid:{BD54C901-076B-434E-B6C7-17C531F4AB41)")', '2026-03-02 13:23:51.350343', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (984, 'Windows: SeDebugPrivilege Enabled by a Suspicious Process', 1, 3, 2, 'Privilege Escalation', 'T1134 - Access Token Manipulation', 'Identifies the creation of a process running as SYSTEM and impersonating a Windows core binary privileges. Adversaries may create a new process with a different token to escalate privileges and bypass access controls.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1134/"]', 'regexMatch("log.action", "(Token Right Adjusted Events)") && regexMatch("log.eventProvider", "(Microsoft-Windows-Security-Auditing)") && regexMatch("log.eventDataEnabledPrivilegeList", "(SeDebugPrivilege)") && !(oneOf("log.eventDataSubjectUserSid", ["S-1-5-18", "S-1-5-19", "S-1-5-20"])) && !(regexMatch("log.eventDataProcessName", "(:\\Windows\\System32\\msiexec.exe|:\\Windows\\SysWOW64\\msiexec.exe|:\\Windows\\System32\\lsass.exe|:\\Windows\\WinSxS\\|:\\Program Files\\|:\\Program Files (x86)\\|:\\Windows\\System32\\MRT.exe|:\\Windows\\System32\\cleanmgr.exe|:\\Windows\\System32\\taskhostw.exe|:\\Windows\\System32\\mmc.exe|:\\Users\\(.+)\\AppData\\Local\\Temp\\(.+)-(.+)\\DismHost.exe|:\\Windows\\System32\\auditpol.exe|:\\Windows\\System32\\wbem\\WmiPrvSe.exe|:\\Windows\\SysWOW64\\wbem\\WmiPrvSe.exe)"))', '2026-03-02 13:23:52.661115', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (985, 'Windows: Suspicious WMIC XSL Script Execution', 2, 3, 2, 'Defense Evasion', 'T1220 - XSL Script Processing', 'Identifies WMIC allowlist bypass techniques by alerting on suspicious execution of scripts. When WMIC loads scripting libraries it may be indicative of an allowlist bypass.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1220/"]', 'regexMatch("log.message", "(?i)wmic.*format") && !contains("log.message", "/format:table") && regexMatch("log.message", "(?i)(jscript\\.dll|vbscript\\.dll)")', '2026-03-02 13:23:54.056297', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (986, 'Windows: Microsoft Exchange Server UM Writing Suspicious Files', 2, 3, 2, 'Initial Access', 'T1190 - Exploit Public-Facing Application', 'Identifies suspicious files being written by the Microsoft Exchange Server Unified Messaging (UM) service. This activity has been observed exploiting CVE-2021-26858.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1190/"]', 'regexMatch("log.eventDataProcessName", "(UMWorkerProcess.exe|umservice.exe)") && regexMatch("log.message", "(php|jsp|js|aspx|asmx|asax|cfm|shtml)") && regexMatch("log.message", "(:\\\\inetpub\\\\wwwroot\\\\aspnet_client\\\\|:\\\\(.+)\\\\Microsoft\\\\Exchange Server(.+)\\\\FrontEnd\\\\HttpProxy\\\\owa\\\\auth\\\\)") && !regexMatch("log.message", "(:\\\\(.+)\\\\Microsoft\\\\Exchange Server(.+)\\\\FrontEnd\\\\HttpProxy\\\\(owa|ecp)\\\\auth\\\\version\\\\)")', '2026-03-02 13:23:55.362289', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (987, 'Windows: Suspicious WMI Image Load from MS Office', 1, 2, 3, 'Execution', 'T1047 - Windows Management Instrumentation', 'Identifies a suspicious image load (wmiutils.dll) from Microsoft Office processes. This behavior may indicate adversarial activity where child processes are spawned via Windows Management Instrumentation (WMI). This technique can be used to execute code and evade traditional parent/child processes spawned from Microsoft Office products.', '["https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/techniques/T1047/"]', 'regexMatch("log.eventDataProcessName", "(WINWORD.EXE|EXCEL.EXE|POWERPNT.EXE|MSPUB.EXE|MSACCESS.EXE)") && contains("log.message", "wmiutils.dll")', '2026-03-02 13:23:56.668790', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (988, 'Windows: Suspicious Image Load (taskschd.dll) from MS Office', 1, 2, 3, 'Persistence', 'T1053 - Scheduled Task/Job', 'Identifies a suspicious image load (taskschd.dll) from Microsoft Office processes. This behavior may indicate adversarial activity where a scheduled task is configured via Windows Component Object Model (COM). This technique can be used to configure persistence and evade monitoring by avoiding the usage of the traditional Windows binary (schtasks.exe) used to manage scheduled tasks.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1053/"]', 'regexMatch("log.eventDataProcessName", "(WINWORD.EXE|EXCEL.EXE|POWERPNT.EXE|MSPUB.EXE|MSACCESS.EXE)") && contains("log.message", "taskschd.dll")', '2026-03-02 13:23:57.811936', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (989, 'Windows: Suspicious Execution from mounted device', 1, 2, 3, 'Defense Evasion', 'T1055 - Process Injection', 'Identifies suspicious process access events from an unknown memory region. Endpoint security solutions usually hook userland Windows APIs in order to decide if the code that is being executed is malicious or not. It''s possible to bypass hooked functions by writing malicious functions that call syscalls directly.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1055/"]', 'equals("log.eventType", "start") && regexMatch("log.process.executable", "(^C:\\\\)") && regexMatch("log.processWorkingDirectory", "((^\\w:\\\\)") && !regexMatch("log.processWorkingDirectory", "(^C:\\\\)") && contains("log.processParentName", "explorer.exe") && oneOf("log.processName", ["rundll32.exe", "mshta.exe", "powershell.exe", "pwsh.exe", "cmd.exe", "regsvr32.exe", "cscript.exe", "wscript.exe"])', '2026-03-02 13:23:59.040429', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (990, 'Windows: New Windows Service Created to start from windows root path. Suspicious event as the binary may have been dropped using Windows Admin Shares', 1, 2, 3, 'Execution', 'T1021.002 - Remote Services: SMB/Windows Admin Shares', 'Adversaries may use Valid Accounts to interact with a remote network share using Server Message Block (SMB).  The adversary may then perform actions as the logged-on user.', '["https://attack.mitre.org/techniques/T1021/002/"]', 'regexMatch("log.eventDataImagePath", "(^%systemroot%\\(.+)\\(.+).exe)") && equals("log.eventCode", 7045) && oneOf("log.logName", ["system", "System"])', '2026-03-02 13:24:00.306277', true, true, 'target', null, '[]', '["target.host","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (991, 'Windows: Suspicious Cmd Execution via WMI', 1, 2, 3, 'Execution', 'T1047 - Windows Management Instrumentation', 'Identifies suspicious command execution (cmd) via Windows Management Instrumentation (WMI) on a remote host. This could be indicative of adversary lateral movement.', '["https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/techniques/T1047/"]', 'contains("log.eventDataParentProcessName", "WmiPrvSE.exe") && contains("log.eventDataProcessName", "cmd.exe") && contains("log.message", "\\\\127.0.0.1\\") && regexMatch("log.message", "(2>&1|1>)")', '2026-03-02 13:24:01.617379', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (992, 'Windows: Suspicious CertUtil Commands', 3, 2, 1, 'Defense Evasion', 'T1140 - Deobfuscate/Decode Files or Information', 'Identifies suspicious commands being used with certutil.exe. CertUtil is a native Windows component which is part of Certificate Services. CertUtil is often abused by attackers to live off the land for stealthier command and control or data exfiltration.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1140/"]', 'regexMatch("log.message", "(decode|encode|urlcache|verifyctl|encodehex|decodehex|exportPFX)") && contains("log.eventDataProcessName", "certutil.exe")', '2026-03-02 13:24:03.011582', true, true, 'target', null, '[]', '["target.host","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (993, 'Windows: Detection of SUNBURST command and control activity', 3, 3, 2, 'Command and Control', 'T1195 - Supply Chain Compromise', 'This rule detects post-exploitation command and control activity of the SUNBURST backdoor.', '["https://attack.mitre.org/tactics/TA0011/","https://attack.mitre.org/techniques/T1195/"]', 'regexMatch("log.eventDataProcessName", "(ConfigurationWizard.exe|NetFlowService.exe|NetflowDatabaseMaintenance.exe|SolarWinds.Administration.exe|SolarWinds.BusinessLayerHost.exe|SolarWinds.BusinessLayerHostx64.exe|SolarWinds.Collector.Service.exe|SolarwindsDiagnostics.exe)") && regexMatch("log.message", "(/swip/Upload.ashx(.+)(POST|PUT)|(POST|PUT)(.+)/swip/Upload.ashx|/swip/SystemDescription(.+)(GET|HEAD)|(GET|HEAD)(.+)/swip/SystemDescription)")', '2026-03-02 13:24:04.375344', true, true, 'target', null, '[]', '["target.host","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (994, 'Windows: Startup Persistence by a Suspicious Process', 1, 2, 3, 'Persistence', 'T1547.001 - Boot or Logon Autostart Execution: Registry Run Keys / Startup Folder', 'Identifies files written to or modified in the startup folder by commonly abused processes. Adversaries may use this technique to maintain persistence.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1547/001/"]', 'contains("log.message", "delet") && regexMatch("log.eventDataProcessName", "(cmd.exe|powershell.exe|wmic.exe|mshta.exe|pwsh.exe|cscript.exe|wscript.exe|regsvr32.exe|RegAsm.exe|rundll32.exe|EQNEDT32.EXE|WINWORD.EXE|EXCEL.EXE|POWERPNT.EXE|MSPUB.EXE|MSACCESS.EXE|iexplore.exe|InstallUtil.exe)") && regexMatch("log.message", "(C:\\Users\\(.+)\\AppData\\Roaming\\Microsoft\\Windows\\Start Menu\\Programs\\Startup\\|C:\\ProgramData\\Microsoft\\Windows\\Start Menu\\Programs\\StartUp\\)") && !contains("target.domain", "NT AUTHORITY")', '2026-03-02 13:24:05.635651', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (995, 'Windows: User account exposed to Kerberoasting', 3, 3, 2, 'Credential Access', 'T1558.003 - Steal or Forge Kerberos Tickets: Kerberoasting', 'Detects when a user account has the servicePrincipalName attribute modified. Attackers can abuse write privileges over a user to configure Service Principle Names (SPNs) so that they can perform Kerberoasting. Administrators can also configure this for legitimate purposes, exposing the account to Kerberoasting.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1558/003/"]', 'regexMatch("log.action", "([Dd]irectory [Ss]ervice [Cc]hanges)") && equals("log.eventCode", 5136) && regexMatch("log.message", "(servicePrincipalName)")', '2026-03-02 13:24:06.774754', true, true, 'target', null, '[]', '["target.host","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (996, 'Windows: SIP Provider Modification', 1, 2, 3, 'Defense Evasion', 'T1553.003 - Subvert Trust Controls: SIP and Trust Provider Hijacking', 'Identifies modifications to the registered Subject Interface Package (SIP) providers. SIP providers are used by the Windows cryptographic system to validate file signatures on the system. This may be an attempt to bypass signature validation checks or inject code into critical processes.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1553/003/"]', 'equals("log.event.type", "change") && contains("log.registryDataStrings", ".dll") && regexMatch("log.registryPath", "(HKLM\\\\SOFTWARE(\\\\WOW6432Node)?\\\\Microsoft\\\\Cryptography\\\\OID\\\\EncodingType 0\\\\CryptSIPDllPutSignedDataMsg\\\\(.+)\\\\Dll|HKLM\\\\SOFTWARE(\\\\WOW6432Node)?\\\\Microsoft\\\\Cryptography\\\\Providers\\\\Trust\\\\FinalPolicy\\\\(.+)\\\\\\$Dll)")', '2026-03-02 13:24:07.824543', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (997, 'Windows: Execution via local SxS Shared Modules', 1, 2, 3, 'Execution', 'T1129 - Shared Modules', 'Identifies the creation, change, or deletion of a DLL module within a Windows SxS local folder. Adversaries may abuse shared modules to execute malicious payloads by instructing the Windows module loader to load DLLs from arbitrary local paths.', '["https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/techniques/T1129/"]', 'contains("log.file.extension", "dll") && regexMatch("log.file.path", "(C:\\\\(.+)\\\\.exe.local\\\\(.+).dll)")', '2026-03-02 13:24:09.001092', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (998, 'Windows: Potential Shadow Credentials added to AD Object', 3, 3, 2, 'Credential Access', 'T1556 - Modify Authentication Process', 'Identify the modification of the msDS-KeyCredentialLink attribute in an Active Directory Computer or User Object. Attackers can abuse control over the object and create a key pair, append to raw public key in the attribute, and obtain persistent and stealthy access to the target user or computer object.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1556/"]', 'regexMatch("log.action", "([Dd]irectory [Ss]ervice [Cc]hanges)") && equals("log.eventCode", 5136) && regexMatch("log.message", "(msDS-KeyCredentialLink)") && contains("log.message", ":828")', '2026-03-02 13:24:10.140157', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (999, 'Windows: Detection of a suspicious PowerShell Script with screenshot capabilities', 3, 2, 1, 'Collection', 'T1113 - Screen Capture', 'Detects PowerShell scripts that can take screenshots, which is a common feature in post-exploitation kits and remote access tools', '["https://attack.mitre.org/tactics/TA0009/","https://attack.mitre.org/techniques/T1113/"]', 'regexMatch("log.message", "(CopyFromScreen(.+)Drawing.Bitmap|Drawing.Bitmap(.+)CopyFromScreen)")', '2026-03-02 13:24:11.277803', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1000, 'Windows: Privilege Escalation via Windir Environment Variable', 2, 3, 2, 'Privilege Escalation', 'T1574.007 - Hijack Execution Flow: Path Interception by PATH Environment Variable', 'Identifies a privilege escalation attempt via a rogue Windows directory (Windir) environment variable. This is a known primitive that is often combined with other vulnerabilities to elevate privileges.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1574/007/"]', 'regexMatch("log.message", "(C:\\windows|%SystemRoot%)") && regexMatch("log.message", "(HKEY_USERS\\(.+)\\Environment\\windir|HKEY_USERS\\(.+)\\Environment\\systemroot)")', '2026-03-02 13:24:12.579516', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1001, 'Windows: Persistence via PowerShell profile', 2, 3, 1, 'Persistence', 'T1546.013 - Event Triggered Execution: PowerShell Profile', 'Identifies the creation or modification of a PowerShell profile. PowerShell profile is a script that is executed when PowerShell starts to customize the user environment, which can be abused by attackers to persist in a environment where PowerShell is common.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1098/002/"]', 'regexMatch("log.eventDataProcessName", "(:\\Users\\(.+)\\Documents\\WindowsPowerShell\\|:\\Users\\(.+)\\Documents\\PowerShell\\|:\\Windows\\System32\\WindowsPowerShell\\)") && regexMatch("log.eventDataProcessName", "(profile.ps1|Microsoft.Powershell_profile.ps1)")', '2026-03-02 13:24:13.722900', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1002, 'Windows: New ActiveSyncAllowedDeviceID Added via PowerShell', 3, 2, 1, 'Persistence', 'T1098.002 - Account Manipulation: Additional Email Delegate Permissions', 'Identifies the use of the Exchange PowerShell cmdlet, Set-CASMailbox, to add a new ActiveSync allowed device. Adversaries may target user email to collect sensitive information.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1098/002/"]', 'oneOf("log.eventDataProcessName", ["powershell.exe", "pwsh.exe", "powershell_ise.exe"]) && regexMatch("log.message", "(Set-CASMailbox(.+)ActiveSyncAllowedDeviceIDs)")', '2026-03-02 13:24:14.906311', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1003, 'Windows: Potential Credential Access via DuplicateHandle in LSASS', 3, 1, 2, 'Credential Access', 'T1003 - OS Credential Dumping', 'Identifies suspicious access to an LSASS handle via DuplicateHandle. This may indicate an attempt to bypass the NtOpenProcess API to evade detection and dump LSASS memory for credential access.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/"]', 'equals("log.eventCode", 10) && contains("log.eventDataProcessName", "lsass.exe") && equals("log.eventDataGrantedAccess", "0x40") && regexMatch("log.eventDataCallTrace", "(UNKNOWN)")', '2026-03-02 13:24:16.167729', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1004, 'Windows: Potential DNS Tunneling via NsLookup', 3, 2, 1, 'Command and Control', 'T1071 - Application Layer Protocol', 'This rule identifies a large number of nslookup.exe executions with an explicit query type from the same host. This may indicate command and control activity utilizing the DNS protocol.', '["https://attack.mitre.org/tactics/TA0011/","https://attack.mitre.org/techniques/T1071/"]', 'contains("log.eventDataProcessName", "nslookup.exe") && regexMatch("log.message", "(-querytype=|-qt=|-q=|-type=)")', '2026-03-02 13:24:17.303228', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1005, 'Windows: Printer driver failed to load, possible remote code execution using PrinterNightmare exploit: CVE-2021-34527', 3, 2, 1, 'Lateral Movement', 'T1210 - Exploitation of Remote Services', 'Adversaries may exploit remote services to gain unauthorized access to internal systems once inside of a network.  Exploitation of a software vulnerability occurs when an adversary takes advantage of a programming error in a program,  service, or within the operating system software or kernel itself to execute adversary-controlled code.  A common goal for post-compromise exploitation of remote services is for lateral movement to enable access to a remote system.', '["https://attack.mitre.org/techniques/T1210/"]', 'equals("log.eventCode", 808) && oneOf("log.severityLabel", ["Error", "error"])', '2026-03-02 13:24:18.699952', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1006, 'Windows: Possible addition of new item to Windows startup registry', 2, 3, 1, 'Persistence', 'T1547.001 - Boot or Logon Autostart Execution: Registry Run Keys / Startup Folder', 'Adversaries may achieve persistence by adding a program to a startup folder or referencing it with a Registry run key.  Adding an entry to the run keys in the Registry or startup folder will cause the program referenced to be executed when a user logs in.  These programs will be executed under the context of the user and will have the account''s associated permissions level.', '["https://attack.mitre.org/techniques/T1547/001/"]', 'regexMatch("log.message", "(SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run)")', '2026-03-02 13:24:19.964043', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1007, 'Windows: PowerShell Script with Token Impersonation Capabilities', 3, 3, 2, 'Privilege Escalation', 'Token Impersonation/Theft', 'Detects scripts that contain PowerShell functions, structures, or Windows API functions related to token impersonation/theft. Attackers may duplicate then impersonate another user''s token to escalate privileges and bypass access controls.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1134/001/"]', e'oneOf("log.message", ["Invoke-TokenManipulation", "ImpersonateNamedPipeClient", "NtImpersonateThread"]) ||
(regexMatch("log.message", "(UpdateProcThreadAttribute(.+)STARTUPINFOEX|STARTUPINFOEX(.+)UpdateProcThreadAttribute)") &&
regexMatch("log.message", "(AdjustTokenPrivileges(.+)SeDebugPrivilege|SeDebugPrivilege(.+)AdjustTokenPrivileges)")) ||
regexMatch("log.message", "((SetThreadToken|ImpersonateLoggedOnUser|CreateProcessWithTokenW|CreatePRocessAsUserW|CreateProcessAsUserA)(.+)(DuplicateToken|DuplicateTokenEx)|(DuplicateToken|DuplicateTokenEx)(.+)(SetThreadToken|ImpersonateLoggedOnUser|CreateProcessWithTokenW|CreatePRocessAsUserW|CreateProcessAsUserA))")
', '2026-03-02 13:24:21.279978', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1008, 'Windows: PowerShell Suspicious Discovery Related Windows API Functions', 2, 1, 3, 'Discovery', 'T1069 - Permission Groups Discovery', 'This rule detects the use of discovery-related Windows API functions in PowerShell Scripts. Attackers can use these functions to perform various situational awareness related activities, like enumerating users, shares, sessions, domain trusts, groups, etc.', '["https://attack.mitre.org/tactics/TA0007/","https://attack.mitre.org/techniques/T1069/"]', e'contains("log.message", ["NetShareEnum", "NetWkstaUserEnum", "NetSessionEnum", "NetLocalGroupEnum", "NetLocalGroupGetMembers", "DsGetSiteName", "DsEnumerateDomainTrusts", "WTSEnumerateSessionsEx", "WTSQuerySessionInformation", "LsaGetLogonSessionData", "QueryServiceObjectSecurity"])
', '2026-03-02 13:24:22.583987', true, true, 'origin', '["adversary.user","adversary.ip"]', '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1009, 'Windows: PowerShell Kerberos Ticket Request', 3, 2, 1, 'Credential Access', 'T1059 - Command and Scripting Interpreter', 'Detects PowerShell scripts that have the capability of requesting kerberos tickets, which is a common step in Kerberoasting toolkits to crack service accounts.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1059/"]', 'regexMatch("log.message", "(KerberosRequestorSecurityToken)")', '2026-03-02 13:24:23.943734', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1010, 'Windows: PowerShell PSReflect Script', 2, 3, 1, 'Execution', 'T1059 - Command and Scripting Interpreter', 'Detects the use of PSReflect in PowerShell scripts. Attackers leverage PSReflect as a library that enables PowerShell to access win32 API functions.', '["https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/techniques/T1059/"]', 'regexMatch("log.message", "(New-InMemoryModule|Add-Win32Type|psenum|DefineDynamicAssembly|DefineDynamicModule|Reflection.TypeAttributes|Reflection.Emit.OpCodes|Reflection.Emit.CustomAttributeBuilder|Runtime.InteropServices.DllImportAttribute)")', '2026-03-02 13:24:25.245292', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1011, 'Windows: Potential Process Injection via PowerShell', 3, 3, 2, 'Defense Evasion', 'T1055 - Process Injection', 'Detects the use of Windows API functions that are commonly abused by malware and security tools to load malicious code or inject it into remote processes.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1055/"]', 'regexMatch("log.message", "(WriteProcessMemory|CreateRemoteThread|NtCreateThreadEx|CreateThread|QueueUserAPC|SuspendThread|ResumeThread|GetDelegateForFunctionPointer)") && regexMatch("log.message", "(VirtualAlloc|VirtualAllocEx|VirtualProtect|LdrLoadDll|LoadLibrary|LoadLibraryA|LoadLibraryEx|GetProcAddress|OpenProcess|OpenProcessToken|AdjustTokenPrivileges)")', '2026-03-02 13:24:26.602880', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1012, 'Windows: Suspicious Portable Executable Encoded in Powershell Script', 2, 3, 1, 'Execution', 'T1059 - Command and Scripting Interpreter', 'Detects the presence of a portable executable (PE) in a PowerShell script by looking for its encoded header. Attackers embed PEs into PowerShell scripts to inject them into memory, avoiding defences by not writing to disk.', '["https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/techniques/T1059/"]', 'contains("log.message", "TVqQAAMAAAAEAAAA")', '2026-03-02 13:24:27.965317', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1013, 'Windows: PowerShell MiniDump Script', 3, 3, 2, 'Credential Access', 'T1059.001 - Command and Scripting Interpreter: PowerShell', 'This rule detects PowerShell scripts capable of dumping process memory using WindowsErrorReporting or Dbghelp.dll MiniDumpWriteDump. Attackers can use this tooling to dump LSASS and get access to credentials.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1059/001/"]', 'regexMatch("log.message", "(MiniDumpWriteDump|MiniDumpWithFullMemory|pmuDetirWpmuDiniM)")', '2026-03-02 13:24:29.271792', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1014, 'Windows: PowerShell Share Enumeration Script', 3, 2, 1, 'Discovery', 'T1135 - Network Share Discovery', 'Detects scripts that contain PowerShell functions, structures, or Windows API functions related to windows share enumeration activities. Attackers, mainly ransomware groups, commonly identify and inspect network shares, looking for critical information for encryption and/or exfiltration.', '["https://attack.mitre.org/tactics/TA0007/","https://attack.mitre.org/techniques/T1135/"]', 'regexMatch("log.message", "(Invoke-ShareFinder|Invoke-ShareFinderThreaded|(shi1_netname(.+)shi1_remark)|shi1_remark(.+)shi1_netname|(NetShareEnum(.+)NetApiBufferFree)|(NetApiBufferFree(.+)NetShareEnum))")', '2026-03-02 13:24:30.360898', true, true, 'origin', '["adversary.user","adversary.ip"]', '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1015, 'Windows: PowerShell Suspicious Payload Encoded and Compressed', 2, 3, 1, 'Defense Evasion', 'T1027 - Obfuscated Files or Information', 'Identifies the use of .NET functionality for decompression and base64 decoding combined in PowerShell scripts, which malware and security tools heavily use to deobfuscate payloads and load them directly in memory to bypass defenses.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1027/"]', 'regexMatch("log.message", "(FromBase64String)") && regexMatch("log.message", "(System.IO.Compression.DeflateStream|System.IO.Compression.GzipStream|IO.Compression.DeflateStream|IO.Compression.GzipStream)")', '2026-03-02 13:24:31.673759', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1016, 'Windows: Suspicious .NET Reflection via PowerShell', 3, 2, 1, 'Defense Evasion', 'T1055 - Process Injection', 'Detects the use of Reflection.Assembly to load PEs and DLLs in memory in PowerShell scripts. Attackers use this method to load executables and DLLs without writing to the disk, bypassing security solutions.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1055/"]', 'regexMatch("log.message", "(\\[Reflection.Assembly\\]::Load)")', '2026-03-02 13:24:32.937200', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1017, 'Windows: Outlook add-in was loaded by powershell, possible use for email collection', 3, 2, 1, 'Execution', 'T1114.001 - Email Collection: Local Email Collection', 'Adversaries may target user email on local systems to collect sensitive information.  Files containing email data can be acquired from a user is local system, such as Outlook storage or cache files.', '["https://attack.mitre.org/techniques/T1114/001/"]', 'contains("log.message", "Microsoft.Office.Interop.Outlook")', '2026-03-02 13:24:34.301901', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1018, 'Windows: Potential Invoke-Mimikatz PowerShell Script', 3, 3, 2, 'Credential Access', 'T1003.001 - OS Credential Dumping: LSASS Memory', 'Mimikatz is a credential dumper capable of obtaining plaintext Windows account logins and passwords, along with many other features that make it useful for testing the security of networks. This rule detects Invoke-Mimikatz PowerShell script and alike.', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/","https://attack.mitre.org/techniques/T1003/001/"]', 'regexMatch("log.message", "(DumpCreds(.+)DumpCerts|DumpCerts(.+)DumpCreds|sekurlsa::logonpasswords|crypto::certificates(.+)CERT_SYSTEM_STORE_LOCAL_MACHINE|CERT_SYSTEM_STORE_LOCAL_MACHINE(.+)crypto::certificates)")', '2026-03-02 13:24:35.609018', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1019, 'Windows: Kerberos Pre-authentication Disabled for User', 2, 2, 3, 'Credential Access', 'T1558.004 - Steal or Forge Kerberos Tickets: AS-REP Roasting', 'Identifies the modification of an accounts Kerberos pre-authentication options. An adversary with GenericWrite/GenericAll rights over the account can maliciously modify these settings to perform offline password cracking attacks such as AS-REP roasting', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1558/","https://attack.mitre.org/techniques/T1558/004/"]', 'regexMatch("log.message", "(('')?[Dd]on('')?t [Rr]equire [Pp]reauth('')?(\\s)?(-)?(\\s)?[Ee]nabled)") && equals("log.eventCode", 4738)', '2026-03-02 13:24:36.922121', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1020, 'Windows audit log was cleared', 1, 2, 3, 'Defense Evasion', 'T1070.001 - Indicator Removal: Clear Windows Event Logs', 'Detects when the Windows audit log (Security event log) has been cleared. Adversaries may clear event logs to remove evidence of an intrusion.', '["https://attack.mitre.org/techniques/T1070/001/"]', 'equals("log.eventCode", 1102)', '2026-03-02 13:24:38.237059', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1021, 'Windows Defender: Protection Disabled', 1, 2, 3, 'Defense Evasion', 'T1562 - Impair Defenses', 'This rule is triggered when it detects that Windows Defender protection has been turned off or disabled on the system. The alert is crucial to identify any unauthorized or malicious actions that may leave the system vulnerable to security threats.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/"]', 'oneOf("log.eventCode", [5001, 5012])', '2026-03-02 13:24:39.501400', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1022, 'Windows: An attempt was made to set the Directory Services Restore Mode', 1, 2, 3, 'Collection', 'T1005 - Data from Local System', 'This event generates on every attempt made to set the Directory Services Restore Mode. Is a function on Active Directory Domain Controllers to take the server offline for emergency maintenance,  particularly restoring backups of AD objects. It is accessed on Windows Server via the advanced startup menu,  similarly to safe mode.', '["https://attack.mitre.org/tactics/TA0009/","https://attack.mitre.org/techniques/T1005"]', 'equals("log.eventCode", 4794)', '2026-03-02 13:24:40.723515', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1023, 'Windows: Hiding files with attrib.exe', 1, 2, 3, 'Defense Evasion', 'T1564 - Hide Artifacts', 'This correlation rule detects events where the attrib.exe tool is used to hide files on the system.  Cloaking files using this tool can be an indicator of suspicious or malicious activity,  as it could be used by malicious actors to evade detection and hide their presence on the system.', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1564/"]', 'regexMatch("log.message", "(attrib +h|attrib.exe +h)")', '2026-03-02 13:24:41.975963', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1024, 'Windows: Changes to AdminSDHolder', 1, 2, 3, 'Privilege Escalation', 'T1087 - Account Discovery', 'This rule detects changes to the AdminSDHolder object within the Active Directory.  The AdminSDHolder object is critical to maintaining the ACLs of highly privileged accounts and groups in the domain.  Any unexpected modification to this object could indicate a privilege escalation attempt or malicious action.', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1069/"]', 'regexMatch("log.action", "Directory Service Changes") && equals("log.eventCode", 5136) && regexMatch("log.eventDataObjectName", "AdminSDHolder")', '2026-03-02 13:24:43.258258', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1074, 'Process Masquerading Detection', 2, 3, 2, 'Defense Evasion', 'T1036.005 - Masquerading: Match Legitimate Name or Location', e'Detects executables masquerading as legitimate Windows system processes but running from
incorrect locations. For example, svchost.exe should only run from C:\\Windows\\System32,
and explorer.exe should only run from C:\\Windows. Malware commonly uses legitimate process
names to avoid detection by analysts and automated tools.

Next Steps:
1. Identify the actual file path of the masquerading process
2. Compare the file hash against known good versions of the legitimate binary
3. Check the digital signature of the suspicious executable
4. Analyze the executable in a sandbox environment
5. Review the parent process that launched the masquerading binary
6. Kill the suspicious process and quarantine the file
7. Search for other instances of the same file across the environment
', '["https://attack.mitre.org/techniques/T1036/005/","https://www.elastic.co/blog/how-hunt-masquerade-ball","https://redcanary.com/threat-detection-report/techniques/masquerading/"]', e'(equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
(
  (
    regexMatch("log.eventDataNewProcessName", "(?i)\\\\\\\\svchost\\\\.exe$") &&
    !regexMatch("log.eventDataNewProcessName", "(?i)^C:\\\\\\\\Windows\\\\\\\\(System32|SysWOW64)\\\\\\\\svchost\\\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataNewProcessName", "(?i)\\\\\\\\csrss\\\\.exe$") &&
    !regexMatch("log.eventDataNewProcessName", "(?i)^C:\\\\\\\\Windows\\\\\\\\(System32|SysWOW64)\\\\\\\\csrss\\\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataNewProcessName", "(?i)\\\\\\\\lsass\\\\.exe$") &&
    !regexMatch("log.eventDataNewProcessName", "(?i)^C:\\\\\\\\Windows\\\\\\\\System32\\\\\\\\lsass\\\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataNewProcessName", "(?i)\\\\\\\\services\\\\.exe$") &&
    !regexMatch("log.eventDataNewProcessName", "(?i)^C:\\\\\\\\Windows\\\\\\\\System32\\\\\\\\services\\\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataNewProcessName", "(?i)\\\\\\\\smss\\\\.exe$") &&
    !regexMatch("log.eventDataNewProcessName", "(?i)^C:\\\\\\\\Windows\\\\\\\\System32\\\\\\\\smss\\\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataNewProcessName", "(?i)\\\\\\\\wininit\\\\.exe$") &&
    !regexMatch("log.eventDataNewProcessName", "(?i)^C:\\\\\\\\Windows\\\\\\\\System32\\\\\\\\wininit\\\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataNewProcessName", "(?i)\\\\\\\\explorer\\\\.exe$") &&
    !regexMatch("log.eventDataNewProcessName", "(?i)^C:\\\\\\\\Windows\\\\\\\\explorer\\\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataProcessName", "(?i)\\\\\\\\svchost\\\\.exe$") &&
    !regexMatch("log.eventDataProcessName", "(?i)^C:\\\\\\\\Windows\\\\\\\\(System32|SysWOW64)\\\\\\\\svchost\\\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataProcessName", "(?i)\\\\\\\\lsass\\\\.exe$") &&
    !regexMatch("log.eventDataProcessName", "(?i)^C:\\\\\\\\Windows\\\\\\\\System32\\\\\\\\lsass\\\\.exe$")
  )
)
', '2026-03-02 13:27:45.750434', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1075, 'LSASS Memory Dump Alternatives', 3, 3, 1, 'Credential Access', 'T1003.001 - OS Credential Dumping: LSASS Memory', e'Detects alternatives to Mimikatz for dumping LSASS process memory, including procdump.exe,
comsvcs.dll MiniDump via rundll32, Task Manager LSASS dump, and direct process access to LSASS.
These tools are commonly used by attackers who avoid Mimikatz to extract credentials from memory.

Next Steps:
1. Isolate the affected host immediately to prevent lateral movement
2. Identify the user account and parent process that initiated the dump
3. Check if procdump or comsvcs.dll was used and from what directory
4. Review for any exfiltration of the dump file to external destinations
5. Reset all credentials that were potentially exposed on the affected host
6. Investigate how the attacker obtained the privileges needed for LSASS access
7. Search for evidence of credential reuse across the environment
', '["https://attack.mitre.org/techniques/T1003/001/","https://www.microsoft.com/en-us/security/blog/2022/10/05/detecting-and-preventing-lsass-credential-dumping-attacks/","https://redcanary.com/threat-detection-report/techniques/lsass-memory/"]', e'(
  (equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
  (
    (regexMatch("log.eventDataCommandLine", "(?i)procdump.*lsass") ||
     regexMatch("log.eventDataCommandLine", "(?i)procdump.*lsass")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)rundll32.*comsvcs\\\\.dll.*MiniDump") ||
     regexMatch("log.eventDataCommandLine", "(?i)rundll32.*comsvcs\\\\.dll.*MiniDump")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)Out-Minidump.*lsass") ||
     regexMatch("log.eventDataCommandLine", "(?i)Out-Minidump.*lsass")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)taskmgr.*lsass") ||
     regexMatch("log.eventDataCommandLine", "(?i)taskmgr.*lsass")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)sqldumper\\\\.exe.*lsass") ||
     regexMatch("log.eventDataCommandLine", "(?i)sqldumper\\\\.exe.*lsass")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)createdump\\\\.exe.*lsass") ||
     regexMatch("log.eventDataCommandLine", "(?i)createdump\\\\.exe.*lsass")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)rdrleakdiag\\\\.exe.*lsass") ||
     regexMatch("log.eventDataCommandLine", "(?i)rdrleakdiag\\\\.exe.*lsass"))
  )
) ||
(
  equals("log.eventCode", "10") &&
  regexMatch("log.eventDataTargetImage", "(?i)lsass\\\\.exe$") &&
  oneOf("log.eventDataGrantedAccess", ["0x1010", "0x1038", "0x1fffff", "0x1410", "0x143a"])
)
', '2026-03-02 13:27:47.013814', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1025, 'Suspicious Admin Share Access Detection', 3, 3, 2, 'Lateral Movement', 'T1021.002 - Remote Services: SMB/Windows Admin Shares', e'Detects suspicious access to Windows administrative shares (C$, ADMIN$, IPC$) which is
commonly used for lateral movement. While admin share access can be legitimate, attackers
use these shares to copy payloads, execute remote commands, and move laterally across the
network. The rule monitors Event ID 5140 (network share access) and Event ID 5145 (detailed
share access) for admin share connections from non-standard sources.

Next Steps:
1. Identify the source IP and user account accessing the admin share
2. Verify if this is authorized administrative access or IT operations
3. Check what files were read or written to the admin share
4. Review if tools or malware were copied via the share
5. Check for service creation or scheduled task creation on the target
6. Correlate with authentication events from the same source IP
7. Block unauthorized admin share access via Group Policy
', '["https://attack.mitre.org/techniques/T1021/002/","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=5140","https://www.sans.org/blog/detecting-lateral-movement-using-windows-admin-shares/"]', e'(
  equals("log.eventCode", "5140") &&
  equals("log.channel", "Security") &&
  regexMatch("log.eventDataShareName", "(?i)\\\\\\\\\\\\*(C\\\\$|ADMIN\\\\$)$") &&
  exists("log.eventDataIpAddress") &&
  !equals("log.eventDataIpAddress", "::1") &&
  !equals("log.eventDataIpAddress", "127.0.0.1") &&
  !regexMatch("log.eventDataSubjectUserName", "(?i)^(SYSTEM|LOCAL SERVICE|NETWORK SERVICE|\\\\$)") &&
  !regexMatch("log.eventDataSubjectUserName", "(?i)\\\\$$")
) ||
(
  equals("log.eventCode", "5145") &&
  equals("log.channel", "Security") &&
  regexMatch("log.eventDataShareName", "(?i)\\\\\\\\\\\\*(C\\\\$|ADMIN\\\\$)$") &&
  regexMatch("log.eventDataRelativeTargetName", "(?i)\\\\.(exe|dll|bat|cmd|ps1|vbs|hta)$") &&
  exists("log.eventDataIpAddress")
)
', '2026-03-02 13:26:42.834233', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataIpAddress","operator":"filter_term","value":"{{.log.eventDataIpAddress}}"}],"or":null,"within":"now-30m","count":3}]', '["lastEvent.log.eventDataIpAddress","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1026, 'Windows Token Manipulation', 3, 3, 3, 'Defense Evasion, Privilege Escalation', 'Access Token Manipulation', 'Detects potential token manipulation attacks used for privilege escalation. Monitors for suspicious combinations of privileged service calls (4673) and operations on privileged objects (4674) along with special privilege assignments (4672) that include sensitive privileges like SeDebugPrivilege, SeImpersonatePrivilege, or SeTcbPrivilege commonly abused for token manipulation.', '["https://medium.com/palantir/windows-privilege-abuse-auditing-detection-and-defense-3078a403d74e","https://attack.mitre.org/techniques/T1134/"]', 'equals("log.eventCode", "4672") && exists("log.eventDataPrivilegeList") && (contains("log.eventDataPrivilegeList", "SeDebugPrivilege") || contains("log.eventDataPrivilegeList", "SeImpersonatePrivilege") || contains("log.eventDataPrivilegeList", "SeTcbPrivilege") || contains("log.eventDataPrivilegeList", "SeAssignPrimaryTokenPrivilege") || contains("log.eventDataPrivilegeList", "SeLoadDriverPrivilege") || contains("log.eventDataPrivilegeList", "SeRestorePrivilege"))', '2026-03-02 13:26:44.141961', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.host","operator":"filter_term","value":"{{.origin.host}}"},{"field":"log.eventDataTargetLogonId","operator":"filter_term","value":"{{.log.eventDataTargetLogonId}}"},{"field":"log.eventCode","operator":"should_terms","value":"4673,4674"}],"or":null,"within":"now-10m","count":3}]', '["lastEvent.log.eventDataTargetLogonId","lastEvent.target.user","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1027, 'Silver Ticket Attack Detection', 3, 3, 2, 'Credential Access', 'T1558.002 - Steal or Forge Kerberos Tickets: Silver Ticket', e'Detects Silver Ticket attacks where adversaries forge Kerberos TGS tickets using a service
account\'s NTLM hash, bypassing the KDC entirely. Unlike Golden Tickets, Silver Tickets target
specific services. The rule detects TGS tickets presented without corresponding TGT requests,
and service access events with anomalous Kerberos authentication patterns.

Next Steps:
1. Identify the targeted service and its associated service account
2. Verify if the service account hash has been compromised via Kerberoasting
3. Reset the targeted service account password immediately
4. Review access logs for the targeted service for unauthorized activity
5. Check for prior Kerberoasting activity targeting the same service SPN
6. Investigate the source host for compromise indicators
7. Implement AES-only encryption for service accounts
8. Enable Kerberos PAC validation on the targeted services
', '["https://attack.mitre.org/techniques/T1558/002/","https://adsecurity.org/?p=2011","https://www.sans.org/blog/kerberos-in-the-crosshairs-golden-tickets-silver-tickets-mitm-and-more/"]', e'equals("log.eventCode", "4769") &&
equals("log.channel", "Security") &&
(
  equals("log.eventDataTicketEncryptionType", "0x17") &&
  !regexMatch("log.eventDataServiceName", "(?i)(krbtgt|\\\\$$)") &&
  !oneOf("log.eventDataStatus", ["0x0", "0x6"]) &&
  exists("log.eventDataIpAddress")
)
', '2026-03-02 13:26:45.499847', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataIpAddress","operator":"filter_term","value":"{{.log.eventDataIpAddress}}"}],"or":null,"within":"now-15m","count":5}]', '["lastEvent.log.eventDataIpAddress","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1028, 'RDP Brute Force Attack', 3, 2, 2, 'Credential Access', 'T1110.001 - Brute Force: Password Guessing', e'Detects multiple failed RDP login attempts from the same source IP address, indicating a potential brute force attack. This rule monitors Windows Event ID 4625 (failed logon) with focus on network logon types (type 3) which are commonly used for RDP connections. The rule triggers when 10 or more failed attempts occur from the same IP within 15 minutes.

Next Steps:
1. Investigate the source IP address for malicious indicators and geolocation
2. Check if the targeted user accounts are legitimate and active
3. Review successful logons from the same IP after failed attempts
4. Implement IP blocking or rate limiting for the source address
5. Enable account lockout policies if not already configured
6. Consider implementing multi-factor authentication for RDP access
7. Review RDP access logs for any successful connections during the attack timeframe
', '["https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4625","https://attack.mitre.org/techniques/T1110/001/"]', 'equals("log.eventCode", "4625") && equals("log.eventDataLogonType", "3") && exists("log.eventDataIpAddress") && !equals("log.eventDataIpAddress", "-") && !equals("log.eventDataIpAddress", "::1") && !equals("log.eventDataIpAddress", "127.0.0.1")', '2026-03-02 13:26:46.801170', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataIpAddress","operator":"filter_term","value":"{{.log.eventDataIpAddress}}"},{"field":"log.eventCode","operator":"filter_term","value":"4625"},{"field":"log.eventDataLogonType","operator":"filter_term","value":"3"}],"or":null,"within":"now-15m","count":10}]', '["adversary.ip","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1029, 'Process Injection Techniques Detection', 3, 3, 2, 'Defense Evasion, Privilege Escalation', 'T1055 - Process Injection', e'Detects various process injection techniques including CreateRemoteThread, SetWindowsHookEx, and other methods used by malware to inject code into legitimate processes. This rule monitors for suspicious cross-process activities targeting critical Windows processes and detects PowerShell/command line usage of injection APIs.

Next Steps:
1. Investigate the source process and command line arguments for suspicious activity
2. Examine the target process for signs of compromise or abnormal behavior
3. Review process tree and parent-child relationships for the involved processes
4. Check for additional suspicious network connections or file system activities from the same host
5. Analyze memory dumps of target processes if available to confirm injection
6. Review system logs for privilege escalation attempts around the same timeframe
7. Correlate with threat intelligence to identify known malware families using similar techniques
8. Consider isolating the affected system if malicious activity is confirmed
', '["https://attack.mitre.org/techniques/T1055/","https://docs.microsoft.com/en-us/sysinternals/downloads/sysmon","https://www.elastic.co/blog/ten-process-injection-techniques-technical-survey-common-and-trending-process"]', e'(
  equals("log.eventCode", "8") &&
  (
    regexMatch("log.eventDataTargetImage", "(?i)(lsass\\\\.exe|csrss\\\\.exe|services\\\\.exe|on\\\\.exe|svchost\\\\.exe|explorer\\\\.exe)") ||
    (
      regexMatch("log.eventDataSourceImage", "(?i)(powershell\\\\.exe|cmd\\\\.exe|rundll32\\\\.exe|regsvr32\\\\.exe)") &&
      !regexMatch("log.eventDataTargetImage", "(?i)(conhost\\\\.exe)")
    )
  )
) ||
(
  equals("log.eventCode", "10") &&
  regexMatch("log.eventDataTargetImage", "(?i)(lsass\\\\.exe|csrss\\\\.exe|services\\\\.exe)") &&
  oneOf("log.eventDataGrantedAccess", ["0x1F0FFF", "0x1F1FFF", "0x1FFFFF", "0x1F3FFF"]) &&
  !regexMatch("log.eventDataSourceImage", "(?i)(taskmgr\\\\.exe|procexp\\\\.exe|procmon\\\\.exe|svchost\\\\.exe)")
) ||
(
  (equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
  (
    regexMatch("log.eventDataCommandLine", "(?i)VirtualAllocEx") ||
    regexMatch("log.eventDataCommandLine", "(?i)WriteProcessMemory") ||
    regexMatch("log.eventDataCommandLine", "(?i)CreateRemoteThread") ||
    regexMatch("log.eventDataCommandLine", "(?i)NtQueueApcThread") ||
    regexMatch("log.eventDataCommandLine", "(?i)SetWindowsHookEx") ||
    regexMatch("log.eventDataCommandLine", "(?i)RtlCreateUserThread")
  )
) ||
(
  (equals("log.eventCode", "4104") || equals("log.eventId", 4104)) &&
  (
    regexMatch("log.eventDataScriptBlockText", "(?i)\\\\[Kernel32\\\\]::(VirtualAllocEx|WriteProcessMemory|CreateRemoteThread)") ||
    regexMatch("log.eventDataScriptBlockText", "(?i)\\\\[ntdll\\\\]::(NtQueueApcThread|RtlCreateUserThread)") ||
    contains("log.eventDataScriptBlockText", "Invoke-ReflectivePEInjection") ||
    contains("log.eventDataScriptBlockText", "Invoke-ProcessHollowing")
  )
)
', '2026-03-02 13:26:48.067827', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.host","operator":"filter_term","value":"{{.origin.host}}"}],"or":null,"within":"now-15m","count":3}]', '["adversary.host","adversary.process"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1030, 'PetitPotam NTLM Relay Attack Detection', 3, 3, 2, 'Credential Access', 'T1187 - Forced Authentication', e'Detects PetitPotam NTLM relay attacks that abuse EFS RPC calls to coerce domain controller
authentication. The attack forces a DC to authenticate to an attacker-controlled host, enabling
NTLM relay to Active Directory Certificate Services (AD CS) for domain compromise. The rule
monitors Event ID 5145 (network share access) for IPC$ share access to specific EFS-related
named pipes used by PetitPotam.

Next Steps:
1. Identify the source IP attempting the EFS pipe access
2. Check if the source is an authorized system or a potential attacker
3. Verify if AD CS is configured and potentially vulnerable to NTLM relay
4. Apply Microsoft patches for PetitPotam (KB5005413)
5. Enable Extended Protection for Authentication on AD CS
6. Disable NTLM authentication where possible
7. Monitor for certificate enrollment from the relayed authentication
8. Restrict access to EFS RPC endpoints via Windows Firewall rules
', '["https://attack.mitre.org/techniques/T1187/","https://github.com/topotam/PetitPotam","https://msrc.microsoft.com/update-guide/vulnerability/ADV210003"]', e'equals("log.eventCode", "5145") &&
equals("log.channel", "Security") &&
equals("log.eventDataShareName", "\\\\\\\\*\\\\IPC$") &&
(
  contains("log.eventDataRelativeTargetName", "efsrpc") ||
  contains("log.eventDataRelativeTargetName", "lsarpc") ||
  contains("log.eventDataRelativeTargetName", "efsr") ||
  contains("log.eventDataRelativeTargetName", "samr")
)
', '2026-03-02 13:26:49.391972', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataIpAddress","operator":"filter_term","value":"{{.log.eventDataIpAddress}}"}],"or":null,"within":"now-5m","count":3}]', '["lastEvent.log.eventDataIpAddress","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1031, 'Pass-the-Hash Attack Detection', 3, 3, 2, 'Lateral Movement', 'T1550.002 - Use Alternate Authentication Material: Pass the Hash', e'Detects Pass-the-Hash attacks by monitoring for NTLM authentication (Event ID 4624) with
LogonType 9 (NewCredentials) or LogonType 3 (Network) from unusual sources, combined with
the use of Seclogon service. Attackers use stolen NTLM hashes to authenticate without
knowing the plaintext password, commonly through tools like Mimikatz sekurlsa::pth,
Impacket, or CrackMapExec.

Next Steps:
1. Identify the source IP and user account used for the NTLM authentication
2. Verify if the source host should be authenticating with NTLM to this target
3. Check for prior credential dumping activity on the source host
4. Review if the authentication was followed by lateral movement or data access
5. Reset the compromised account password and any related accounts
6. Implement NTLM restrictions via Group Policy where possible
7. Enable Windows Defender Credential Guard to protect NTLM hashes
', '["https://attack.mitre.org/techniques/T1550/002/","https://www.sans.org/blog/pass-the-hash-attack-detection/","https://stealthbits.com/blog/how-to-detect-pass-the-hash-attacks/"]', e'(
  equals("log.eventCode", "4624") &&
  equals("log.channel", "Security") &&
  equals("log.eventDataLogonType", "9") &&
  equals("log.eventDataAuthenticationPackageName", "Negotiate") &&
  !regexMatch("log.eventDataSubjectUserName", "(?i)^(SYSTEM|LOCAL SERVICE|NETWORK SERVICE|ANONYMOUS LOGON|-|\\\\$)") &&
  exists("target.user") &&
  !regexMatch("target.user", "(?i)\\\\$$")
) ||
(
  equals("log.eventCode", "4624") &&
  equals("log.channel", "Security") &&
  equals("log.eventDataLogonType", "3") &&
  equals("log.eventDataLmPackageName", "NTLM V1") &&
  exists("log.eventDataIpAddress") &&
  !equals("log.eventDataIpAddress", "-") &&
  !equals("log.eventDataIpAddress", "::1") &&
  !equals("log.eventDataIpAddress", "127.0.0.1")
)
', '2026-03-02 13:26:50.707553', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataIpAddress","operator":"filter_term","value":"{{.log.eventDataIpAddress}}"},{"field":"log.eventCode","operator":"filter_term","value":"4624"}],"or":null,"within":"now-30m","count":3}]', '["lastEvent.log.eventDataIpAddress","adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1032, 'NTDS.dit Extraction Attempt', 3, 3, 1, 'Credential Access', 'T1003.003 - OS Credential Dumping: NTDS', e'Detects attempts to access or copy the Active Directory domain database (NTDS.dit) which contains password hashes for all domain users. This is a critical indicator of credential theft attempts and potential domain compromise.

Next Steps:
1. Immediately isolate the affected system to prevent further compromise
2. Review all recent activity from the source host and user account
3. Check for signs of lateral movement from this system
4. Verify integrity of domain controllers and examine recent administrative actions
5. Look for evidence of credential harvesting tools (ntdsutil, vssadmin, mimikatz)
6. Review privileged account usage and consider forcing password resets
7. Examine network traffic for data exfiltration attempts
8. Check backup systems and shadow copies for unauthorized access
9. Coordinate with incident response team for full forensic analysis
', '["https://attack.mitre.org/techniques/T1003/003/","https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-10/security/threat-protection/auditing/event-4663"]', e'oneOf("log.eventCode", ["4663", "4656"]) &&
equals("log.channel", "Security") &&
(
  endsWith("log.eventDataObjectName", "\\\\ntds.dit") ||
  contains("log.eventDataObjectName", "\\\\NTDS\\\\") ||
  endsWith("log.eventDataProcessName", "\\\\ntdsutil.exe") ||
  endsWith("log.eventDataProcessName", "\\\\vssadmin.exe")
) &&
!equals("log.eventDataAccessMask", "0x0")
', '2026-03-02 13:26:51.792644', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.host","operator":"filter_term","value":"{{.origin.host}}"}],"or":null,"within":"now-30m","count":2}]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1033, 'Malicious PowerShell Execution', 3, 3, 2, 'Execution', 'T1059.001 - Command and Scripting Interpreter: PowerShell', e'Detects malicious PowerShell execution patterns including encoded commands, download cradles,
AMSI bypass attempts, execution policy bypasses, and common offensive PowerShell techniques.
These patterns are frequently used by attackers for initial access, lateral movement, and
payload delivery.

Next Steps:
1. Isolate the affected host immediately
2. Decode any Base64-encoded command content for analysis
3. Check parent process - unexpected parents (e.g., Word, Excel) indicate macro-based attacks
4. Review network connections for download cradle destinations
5. Examine PowerShell transcription and script block logs for full command content
6. Search for persistence mechanisms created by the script
7. Check if AMSI was successfully bypassed and what payload was executed
8. Correlate with any recent phishing emails received by the user
', '["https://attack.mitre.org/techniques/T1059/001/","https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_logging","https://www.fireeye.com/blog/threat-research/2016/02/greater_visibility_t.html"]', e'(
  (equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
  (
    regexMatch("log.eventDataNewProcessName", "(?i)powershell\\\\.exe$|pwsh\\\\.exe$") ||
    regexMatch("log.eventDataProcessName", "(?i)powershell\\\\.exe$|pwsh\\\\.exe$")
  ) &&
  (
    regexMatch("log.eventDataCommandLine", "(?i)-[eE][nN][cC]\\\\s+[A-Za-z0-9+/=]{50,}") ||
    regexMatch("log.eventDataCommandLine", "(?i)(-nop|-noni|-w\\\\s+hidden|-ep\\\\s+bypass|-exec\\\\s+bypass)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(DownloadString|DownloadFile|DownloadData|WebClient|Invoke-WebRequest|wget|curl|Start-BitsTransfer)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(Invoke-Expression|IEX|Invoke-Command)\\\\s*[\\\\(\\\\{]") ||
    regexMatch("log.eventDataCommandLine", "(?i)(Net\\\\.WebClient|IO\\\\.MemoryStream|IO\\\\.StreamReader|IO\\\\.Compression)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(FromBase64String|ToBase64String|\\\\[Convert\\\\]::)") ||
    regexMatch("log.eventDataCommandLine", "(?i)-[eE][nN][cC]\\\\s+[A-Za-z0-9+/=]{50,}") ||
    regexMatch("log.eventDataCommandLine", "(?i)(-nop|-noni|-w\\\\s+hidden|-ep\\\\s+bypass|-exec\\\\s+bypass)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(DownloadString|DownloadFile|DownloadData|WebClient|Invoke-WebRequest)")
  )
) ||
(
  (equals("log.eventCode", "4104") || equals("log.eventId", 4104)) &&
  equals("log.providerName", "Microsoft-Windows-PowerShell") &&
  (
    contains("log.eventDataScriptBlockText", "AmsiInitFailed") ||
    contains("log.eventDataScriptBlockText", "amsiContext") ||
    contains("log.eventDataScriptBlockText", "AmsiUtils") ||
    contains("log.eventDataScriptBlockText", "amsi.dll") ||
    regexMatch("log.eventDataScriptBlockText", "(?i)(Invoke-Obfuscation|Invoke-CradleCrafter|Out-EncodedCommand)") ||
    regexMatch("log.eventDataScriptBlockText", "(?i)(New-Object\\\\s+Net\\\\.Sockets\\\\.TCPClient|Net\\\\.WebClient\\\\)?\\\\.DownloadString)") ||
    regexMatch("log.eventDataScriptBlockText", "(?i)(\\\\$env:COMSPEC|cmd\\\\.exe.*/c)") ||
    regexMatch("log.eventDataScriptBlockText", "(?i)(Add-Type.*DllImport|\\\\[DllImport)") ||
    contains("log.eventDataScriptBlockText", "Reflection.Assembly") ||
    contains("log.eventDataScriptBlockText", "Invoke-Shellcode") ||
    contains("log.eventDataScriptBlockText", "AmsiInitFailed") ||
    contains("log.eventDataScriptBlockText", "AmsiUtils") ||
    regexMatch("log.eventDataScriptBlockText", "(?i)(Invoke-Obfuscation|Invoke-CradleCrafter|Out-EncodedCommand)")
  )
)
', '2026-03-02 13:26:53.057252', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.host","operator":"filter_term","value":"{{.origin.host}}"}],"or":null,"within":"now-10m","count":2}]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1034, 'Keylogger Activity Detection', 3, 2, 0, 'Collection', 'T1056.001 - Input Capture: Keylogging', e'Detects keylogger activity through Windows API hook installation, clipboard monitoring,
keyboard input capture, and known keylogger tool execution. Attackers use keyloggers
to capture user credentials, sensitive data, and communications.

Next Steps:
1. Immediately isolate the affected host to stop credential capture
2. Identify the source process installing keyboard hooks and its origin
3. Check if the hooking process is a known legitimate application
4. Review what user accounts have been active on the host during the capture period
5. Force password resets for all accounts used on the compromised system
6. Check for data exfiltration - keylog data being sent externally
7. Examine the process tree to find how the keylogger was installed
8. Scan for persistence mechanisms associated with the keylogger
9. Review MFA tokens and session cookies that may have been captured
', '["https://attack.mitre.org/techniques/T1056/001/","https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setwindowshookexw","https://www.sans.org/reading-room/whitepapers/detection/detecting-keyloggers-36062"]', e'(
  (equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
  (
    regexMatch("log.eventDataCommandLine", "(?i)(GetAsyncKeyState|GetKeyState|GetKeyboardState|SetWindowsHookEx|SetWindowsHookExW|SetWindowsHookExA)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(keylog|key.?log|key.?stroke|key.?capture|keystroke.?log)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(Get-Clipboard|Set-Clipboard|clipboard.?monitor|ClipboardChanged)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(GetAsyncKeyState|GetKeyState|SetWindowsHookEx)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(keylog|key.?log|key.?stroke|key.?capture)")
  )
) ||
(
  (equals("log.eventCode", "4104") || equals("log.eventId", 4104)) &&
  equals("log.providerName", "Microsoft-Windows-PowerShell") &&
  (
    contains("log.eventDataScriptBlockText", "GetAsyncKeyState") ||
    contains("log.eventDataScriptBlockText", "GetKeyboardState") ||
    contains("log.eventDataScriptBlockText", "SetWindowsHookEx") ||
    contains("log.eventDataScriptBlockText", "WH_KEYBOARD_LL") ||
    contains("log.eventDataScriptBlockText", "WH_KEYBOARD") ||
    contains("log.eventDataScriptBlockText", "MapVirtualKey") ||
    contains("log.eventDataScriptBlockText", "Get-Keystrokes") ||
    contains("log.eventDataScriptBlockText", "Invoke-Keylogger") ||
    (
      contains("log.eventDataScriptBlockText", "user32.dll") &&
      (
        contains("log.eventDataScriptBlockText", "GetAsyncKeyState") ||
        contains("log.eventDataScriptBlockText", "GetForegroundWindow") ||
        contains("log.eventDataScriptBlockText", "GetWindowText")
      )
    ) ||
    (
      contains("log.eventDataScriptBlockText", "DllImport") &&
      contains("log.eventDataScriptBlockText", "user32") &&
      contains("log.eventDataScriptBlockText", "KeyState")
    ) ||
    contains("log.eventDataScriptBlockText", "OpenClipboard") ||
    contains("log.eventDataScriptBlockText", "GetClipboardData") ||
    contains("log.eventDataScriptBlockText", "GetAsyncKeyState") ||
    contains("log.eventDataScriptBlockText", "SetWindowsHookEx") ||
    contains("log.eventDataScriptBlockText", "WH_KEYBOARD_LL")
  )
) ||
(
  equals("log.eventCode", "8") &&
  contains("log.eventDataStartModule", "user32.dll") &&
  contains("log.eventDataStartFunction", "SetWindowsHookEx")
)
', '2026-03-02 13:26:54.099883', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.host","operator":"filter_term","value":"{{.origin.host}}"}],"or":null,"within":"now-15m","count":2}]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1035, 'Kerberoasting Attack Detection', 3, 2, 1, 'Credential Access', 'T1558.003 - Steal or Forge Kerberos Tickets: Kerberoasting', e'Detects Kerberoasting attacks where adversaries request Kerberos TGS tickets encrypted with RC4 (0x17) for
service accounts in order to crack them offline and obtain plaintext credentials. This is the most common
Active Directory credential theft technique used in real-world compromises. The rule monitors Event ID 4769
(Kerberos Service Ticket Operations) for RC4 encryption requests while excluding machine accounts (ending in $)
and legitimate system services.

Next Steps:
1. Identify the requesting user account and verify if this is authorized security testing
2. Check which service account SPNs were targeted for TGS requests
3. Review if the requesting account has been compromised
4. Audit all service accounts with SPNs for weak passwords
5. Consider implementing AES-only Kerberos encryption policies
6. Rotate passwords for targeted service accounts immediately
7. Enable Group Managed Service Accounts (gMSA) where possible
8. Monitor for follow-up lateral movement using obtained credentials
', '["https://attack.mitre.org/techniques/T1558/003/","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4769","https://adsecurity.org/?p=2293"]', e'equals("log.eventCode", "4769") &&
equals("log.channel", "Security") &&
equals("log.eventDataTicketEncryptionType", "0x17") &&
!regexMatch("log.eventDataServiceName", "(?i)\\\\$$") &&
!equals("log.eventDataServiceName", "krbtgt") &&
!oneOf("log.eventDataTicketOptions", ["0x40810000", "0x40800000", "0x40810010"]) &&
exists("log.eventDataServiceName")
', '2026-03-02 13:26:55.292640', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataIpAddress","operator":"filter_term","value":"{{.log.eventDataIpAddress}}"}],"or":null,"within":"now-15m","count":3}]', '["lastEvent.log.eventDataIpAddress","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1036, 'DCSync Attack Detection', 3, 3, 1, 'Credential Access', 'T1003.006 - OS Credential Dumping: DCSync', e'Detects DCSync attacks where attackers use directory replication services to retrieve password hashes from domain controllers.
This technique exploits legitimate Active Directory replication functionality to extract credentials without directly accessing the domain controller\'s files.
The rule monitors for specific replication GUIDs associated with credential access operations in Windows Event ID 4662.

Next Steps:
- Immediately verify the legitimacy of the user account performing the replication operation
- Check if the source host is an authorized domain controller or backup system
- Review recent privilege escalation activities for the identified user account
- Examine network traffic for additional signs of credential harvesting
- Consider resetting passwords for high-privilege accounts if compromise is confirmed
- Review domain controller access logs for unauthorized administrative activities
', '["https://attack.mitre.org/techniques/T1003/006/","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4662","https://www.elastic.co/guide/en/security/current/potential-credential-access-via-dcsync.html"]', e'equals("log.eventCode", "4662") &&
equals("log.channel", "Security") &&
equals("log.eventDataObjectServer", "DS") &&
(
  contains("log.eventDataProperties", "1131f6aa-9c07-11d1-f79f-00c04fc2dcd2") ||
  contains("log.eventDataProperties", "1131f6ad-9c07-11d1-f79f-00c04fc2dcd2") ||
  contains("log.eventDataProperties", "89e95b76-444d-4c62-991a-0facbeda640c") ||
  contains("log.eventDataProperties", "19195a5b-6da0-11d0-afd3-00c04fd930c9")
) &&
!regexMatch("log.eventDataSubjectUserName", ".*\\\\$$") &&
!contains("origin.host", "DC")
', '2026-03-02 13:26:56.474938', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataSubjectUserName","operator":"filter_term","value":"{{.log.eventDataSubjectUserName}}"}],"or":null,"within":"now-1h","count":1}]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1037, 'DCShadow Attack Detection', 3, 3, 2, 'Defense Evasion', 'T1207 - Rogue Domain Controller', e'Detects DCShadow attacks where attackers register a rogue domain controller to push malicious Active Directory changes. This technique allows adversaries to modify Active Directory objects by registering a rogue domain controller and triggering replication, effectively bypassing security controls and detection mechanisms.

The rule monitors for:
- Computer account modifications with domain controller service principal names
- Access to sensitive Active Directory objects and properties
- Creation of server objects in the domain controller configuration

Next Steps:
1. Immediately investigate the source host and user account involved in the activity
2. Check if the host is an authorized domain controller in your environment
3. Review recent Active Directory changes and replication logs
4. Examine authentication logs for the affected user account
5. Verify the legitimacy of any recent domain controller promotions
6. Check for signs of compromise on the source system
7. Consider isolating the affected host if unauthorized activity is confirmed
8. Review domain controller security policies and access controls
', '["https://attack.mitre.org/techniques/T1207/","https://www.dcshadow.com/","https://blog.alsid.eu/dcshadow-explained-4510f52fc19d"]', e'(
  (equals("log.eventCode", "4742") &&
   equals("log.channel", "Security") &&
   contains("log.eventDataServicePrincipalNames", "GC/") &&
   contains("log.eventDataUserAccountControl", "SERVER_TRUST_ACCOUNT")) ||
  (equals("log.eventCode", "4662") &&
   equals("log.channel", "Security") &&
   equals("log.eventDataObjectType", "{bf967a92-0de6-11d0-a285-00aa003049e2}") &&
   contains("log.eventDataProperties", "1131f6ac-9c07-11d1-f79f-00c04fc2dcd2")) ||
  (equals("log.eventCode", "5137") &&
   equals("log.channel", "Security") &&
   equals("log.eventDataObjectClass", "server") &&
   contains("log.eventDataObjectDN", "CN=Servers,CN="))
) &&
!contains("origin.host", "DC")
', '2026-03-02 13:26:57.632855', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.host","operator":"filter_term","value":"{{.origin.host}}"}],"or":null,"within":"now-2h","count":2}]', '["lastEvent.log.eventDataSubjectUserName","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1038, 'Golden Ticket Attack Detection', 3, 3, 3, 'Credential Access', 'T1558.001 - Steal or Forge Kerberos Tickets: Golden Ticket', e'Detects Golden Ticket attacks where adversaries forge Kerberos TGTs using the KRBTGT account
hash, granting unlimited domain access. The rule detects anomalous TGT usage patterns including
TGS requests with unusual encryption types, tickets with abnormally long lifetimes, and Kerberos
authentication from non-domain-controller sources for the KRBTGT service.

Next Steps:
1. Immediately verify if the KRBTGT account password has been compromised
2. Reset the KRBTGT password TWICE to invalidate all existing tickets
3. Identify the source host and investigate for full domain compromise
4. Review all domain admin activity from the suspected timeframe
5. Check for DCSync or NTDS.dit extraction as precursor activities
6. Audit all privileged account access across the domain
7. Consider rebuilding the domain if compromise is confirmed
8. Implement Kerberos armoring and constrained delegation
', '["https://attack.mitre.org/techniques/T1558/001/","https://adsecurity.org/?p=1640","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4769"]', e'(
  equals("log.eventCode", "4769") &&
  equals("log.channel", "Security") &&
  equals("log.eventDataServiceName", "krbtgt") &&
  !equals("log.eventDataStatus", "0x0") &&
  exists("log.eventDataIpAddress")
) ||
(
  equals("log.eventCode", "4768") &&
  equals("log.channel", "Security") &&
  !oneOf("log.eventDataTicketEncryptionType", ["0x12", "0x11"]) &&
  exists("target.user") &&
  !regexMatch("target.user", "(?i)\\\\$$")
) ||
(
  equals("log.eventCode", "4672") &&
  equals("log.channel", "Security") &&
  contains("log.eventDataPrivilegeList", "SeTcbPrivilege") &&
  !regexMatch("log.eventDataSubjectUserName", "(?i)^(SYSTEM|LOCAL SERVICE|NETWORK SERVICE)$") &&
  !regexMatch("log.eventDataSubjectUserName", "(?i)\\\\$$")
)
', '2026-03-02 13:26:58.868113', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.host","operator":"filter_term","value":"{{.origin.host}}"}],"or":null,"within":"now-30m","count":3}]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1039, 'BloodHound Reconnaissance Activity', 3, 2, 1, 'Discovery', 'T1087 - Account Discovery', e'Detects potential BloodHound Active Directory reconnaissance tool usage through LDAP queries, characteristic patterns, and AD enumeration activities. BloodHound is commonly used by attackers to map Active Directory relationships and identify privilege escalation paths.

Next Steps:
1. Investigate the source host and user account involved in the activity
2. Review network logs for LDAP queries to domain controllers around the same timeframe
3. Check for other reconnaissance tools or suspicious PowerShell activity on the same host
4. Examine Active Directory audit logs for unusual object access patterns
5. Verify if the user account has legitimate reasons for AD enumeration activities
6. Look for signs of lateral movement or privilege escalation following this reconnaissance
7. Consider isolating the affected host if malicious activity is confirmed
', '["https://attack.mitre.org/techniques/T1087/","https://bloodhound.readthedocs.io/","https://docs.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4662"]', e'(
  (equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
  (
    regexMatch("log.eventDataNewProcessName", "(?i)bloodhound") ||
    regexMatch("log.eventDataNewProcessName", "(?i)sharphound") ||
    regexMatch("log.eventDataCommandLine", "(?i)(bloodhound|sharphound)") ||
    regexMatch("log.eventDataCommandLine", "(?i)--CollectionMethod\\\\s+(All|Session|LoggedOn)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(DCOnly|ComputerOnly|LocalGroup)")
  )
) ||
(
  (equals("log.eventCode", "4104") || equals("log.eventId", 4104)) &&
  equals("log.providerName", "Microsoft-Windows-PowerShell") &&
  (
    regexMatch("log.eventDataScriptBlockText", "(?i)invoke-bloodhound") ||
    contains("log.eventDataScriptBlockText", "Get-BloodHoundData") ||
    contains("log.eventDataScriptBlockText", "Get-NetSession") ||
    contains("log.eventDataScriptBlockText", "Get-NetLoggedOn") ||
    contains("log.eventDataScriptBlockText", "Get-DomainTrust")
  )
) ||
(
  equals("log.eventCode", "4662") &&
  regexMatch("log.eventDataObjectType", "(?i)(bf967aba-0de6-11d0-a285-00aa003049e2|bf967a9c-0de6-11d0-a285-00aa003049e2)") &&
  oneOf("log.eventDataAccessMask", ["0x100", "0x10000"])
) ||
(
  equals("log.eventCode", "5156") &&
  equals("log.eventDataDestinationPort", "389") &&
  equals("log.eventDataDirection", "%%14592")
)
', '2026-03-02 13:27:00.135533', true, true, 'origin', '["adversary.host","lastEvent.log.eventDataSubjectUserName"]', '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.host","operator":"filter_term","value":"{{.origin.host}}"}],"or":null,"within":"now-2h","count":10}]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1040, 'AS-REP Roasting Attack Detection', 3, 2, 1, 'Credential Access', 'T1558.004 - Steal or Forge Kerberos Tickets: AS-REP Roasting', e'Detects AS-REP Roasting attacks targeting accounts with Kerberos pre-authentication disabled.
Attackers request AS-REP messages encrypted with RC4 (0x17) for accounts that do not require
pre-authentication, enabling offline password cracking. This is a companion technique to Kerberoasting
and targets a different set of vulnerable accounts.

Next Steps:
1. Identify accounts with pre-authentication disabled and evaluate business justification
2. Enable Kerberos pre-authentication on all identified accounts
3. Verify the requesting source IP is not a known attack tool
4. Reset passwords for targeted accounts using strong, complex passwords
5. Audit Active Directory for accounts with DONT_REQUIRE_PREAUTH flag
6. Monitor for subsequent credential usage from the requesting IP
', '["https://attack.mitre.org/techniques/T1558/004/","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4768","https://blog.harmj0y.net/activedirectory/roasting-as-reps/"]', e'equals("log.eventCode", "4768") &&
equals("log.channel", "Security") &&
equals("log.eventDataTicketEncryptionType", "0x17") &&
equals("log.eventDataPreAuthType", "0") &&
!regexMatch("target.user", "(?i)\\\\$$") &&
exists("target.user")
', '2026-03-02 13:27:01.581886', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataIpAddress","operator":"filter_term","value":"{{.log.eventDataIpAddress}}"}],"or":null,"within":"now-15m","count":3}]', '["lastEvent.log.eventDataIpAddress","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1041, 'ADFS Authentication Anomalies', 3, 2, 2, 'Defense Evasion, Persistence, Privilege Escalation, Initial Access', 'T1078 - Valid Accounts', e'Detects anomalous authentication attempts against Active Directory Federation Services (ADFS) including multiple failed attempts that could indicate password spraying or brute force attacks. This rule monitors for authentication failures, token validation failures, and other ADFS security events that may indicate malicious activity.

Next Steps:
1. Review the source IP address and determine if it\'s from a known/trusted location
2. Check for patterns of failed authentication attempts across multiple users
3. Examine ADFS audit logs for additional context around the authentication failures
4. Verify if the targeted user accounts are valid and active
5. Consider implementing IP-based blocking if malicious activity is confirmed
6. Review ADFS configuration for security hardening opportunities
7. Correlate with other authentication events across the domain
', '["https://learn.microsoft.com/en-us/windows-server/identity/ad-fs/troubleshooting/ad-fs-tshoot-logging","https://attack.mitre.org/techniques/T1078/"]', 'equals("log.providerName", "AD FS") && (equals("log.eventId", "411") || equals("log.eventId", "342") || equals("log.eventId", "516")) && contains("log.message", "token validation failed")', '2026-03-02 13:27:02.890185', true, true, 'origin', null, '[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataIpAddress","operator":"filter_term","value":"{{.log.eventDataIpAddress}}"}],"or":null,"within":"now-10m","count":10}]', '["adversary.ip","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1042, 'WMI Remote Command Execution Detection', 3, 3, 2, 'Lateral Movement', 'T1047 - Windows Management Instrumentation', e'Detects WMI-based remote command execution using wmic.exe /node: parameter or PowerShell
Invoke-WmiMethod / Invoke-CimMethod for lateral movement. Attackers use WMI to execute
commands on remote systems without dropping files to disk, making it a stealthy lateral
movement technique that leverages built-in Windows functionality.

Next Steps:
1. Identify the target host specified in the /node: parameter
2. Verify if this is authorized administrative WMI usage
3. Check what process or command was executed on the remote host
4. Review the credentials used for the WMI connection
5. Search for WMI execution evidence on the target host
6. Check for additional lateral movement from the same source
7. Monitor for process creation events on the target system
', '["https://attack.mitre.org/techniques/T1047/","https://www.fireeye.com/content/dam/fireeye-www/global/en/current-threats/pdfs/wp-windows-management-instrumentation.pdf","https://www.sans.org/blog/wmi-for-detection-and-response/"]', e'(equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
(
  (regexMatch("log.eventDataCommandLine", "(?i)wmic(.exe)?\\\\s+/node:") ||
   regexMatch("log.eventDataCommandLine", "(?i)wmic(.exe)?\\\\s+/node:")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)Invoke-WmiMethod.*-ComputerName") ||
   regexMatch("log.eventDataCommandLine", "(?i)Invoke-WmiMethod.*-ComputerName")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)Invoke-CimMethod.*-ComputerName") ||
   regexMatch("log.eventDataCommandLine", "(?i)Invoke-CimMethod.*-ComputerName")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)Get-WmiObject.*-ComputerName.*Win32_Process") ||
   regexMatch("log.eventDataCommandLine", "(?i)Get-WmiObject.*-ComputerName.*Win32_Process")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)wmic(.exe)?.*process\\\\s+call\\\\s+create") ||
   regexMatch("log.eventDataCommandLine", "(?i)wmic(.exe)?.*process\\\\s+call\\\\s+create"))
)
', '2026-03-02 13:27:04.246580', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1043, 'WMI Event Subscription Persistence Detection', 2, 3, 2, 'Persistence', 'T1546.003 - Event Triggered Execution: Windows Management Instrumentation Event Subscription', e'Detects creation of WMI event subscriptions used for stealthy persistence. Attackers create
WMI event filters and consumers that execute commands when specific system events occur.
This is a highly effective persistence mechanism because it survives reboots and is difficult
to detect without specific monitoring. The rule monitors for wmic.exe commands creating
event subscriptions and PowerShell WMI subscription creation patterns.

Next Steps:
1. Enumerate all WMI event subscriptions using Get-WMIObject or wmic
2. Examine the event filter conditions and consumer actions
3. Identify the command or script executed by the consumer
4. Remove malicious WMI subscriptions immediately
5. Check for related persistence mechanisms on the same host
6. Review the timeline to identify the initial compromise
7. Search for similar WMI persistence across other endpoints
', '["https://attack.mitre.org/techniques/T1546/003/","https://www.fireeye.com/blog/threat-research/2016/08/wmi_vs_wmi_monitor.html","https://pentestlab.blog/2020/01/21/persistence-wmi-event-subscription/"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  (
    regexMatch("log.eventDataCommandLine", "(?i)wmic.*(/namespace|__EventFilter|__EventConsumer|__FilterToConsumerBinding)") ||
    regexMatch("log.eventDataCommandLine", "(?i)wmic.*(/namespace|__EventFilter|__EventConsumer|__FilterToConsumerBinding)")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)(Set-WmiInstance|New-CimInstance).*(__EventFilter|__EventConsumer|CommandLineEventConsumer|ActiveScriptEventConsumer)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(Set-WmiInstance|New-CimInstance).*(__EventFilter|__EventConsumer|CommandLineEventConsumer|ActiveScriptEventConsumer)")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)Register-WmiEvent") ||
    regexMatch("log.eventDataCommandLine", "(?i)Register-WmiEvent")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)mofcomp\\\\.exe") ||
    regexMatch("log.eventDataCommandLine", "(?i)mofcomp\\\\.exe")
  )
) ||
(
  (equals("log.eventCode", "4104") || equals("log.eventId", 4104)) &&
  equals("log.providerName", "Microsoft-Windows-PowerShell") &&
  (
    regexMatch("log.eventDataScriptBlockText", "(?i)(__EventFilter|__EventConsumer|__FilterToConsumerBinding)") ||
    regexMatch("log.eventDataScriptBlockText", "(?i)(CommandLineEventConsumer|ActiveScriptEventConsumer)") ||
    contains("log.eventDataScriptBlockText", "Register-WmiEvent")
  )
)
', '2026-03-02 13:27:05.511230', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1044, 'Windows Script Host Suspicious Execution', 3, 3, 2, 'Execution', 'T1059.005 - Command and Scripting Interpreter: Visual Basic', e'Detects suspicious execution of wscript.exe and cscript.exe with potentially malicious script
files or command-line arguments. Attackers commonly use Windows Script Host to execute VBScript
and JScript payloads delivered via phishing emails, drive-by downloads, or as secondary
payloads during an intrusion. The rule focuses on scripts executed from suspicious locations
and with suspicious arguments.

Next Steps:
1. Identify the script file being executed and analyze its contents
2. Check the parent process (e.g., Outlook, Word, browser) for phishing indicators
3. Review the script file location for legitimacy
4. Analyze any network connections made by the script
5. Check if the script downloads or executes additional payloads
6. Search for the same script across other endpoints
7. Block the script and remove any persistence mechanisms it created
', '["https://attack.mitre.org/techniques/T1059/005/","https://lolbas-project.github.io/lolbas/Binaries/Wscript/","https://lolbas-project.github.io/lolbas/Binaries/Cscript/"]', e'(equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
(
  regexMatch("log.eventDataNewProcessName", "(?i)\\\\\\\\(wscript|cscript)\\\\.exe$") ||
  regexMatch("log.eventDataProcessName", "(?i)\\\\\\\\(wscript|cscript)\\\\.exe$")
) &&
(
  (regexMatch("log.eventDataCommandLine", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\|\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Users\\\\\\\\Public\\\\\\\\)") ||
   regexMatch("log.eventDataCommandLine", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\|\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Users\\\\\\\\Public\\\\\\\\)")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)//[eE]:(vbscript|jscript)") ||
   regexMatch("log.eventDataCommandLine", "(?i)//[eE]:(vbscript|jscript)")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)(http://|https://|ftp://|\\\\\\\\\\\\\\\\[0-9])") ||
   regexMatch("log.eventDataCommandLine", "(?i)(http://|https://|ftp://|\\\\\\\\\\\\\\\\[0-9])")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)\\\\.(vbs|vbe|js|jse|wsf|wsh)\\\\s+(//B|//nologo)") ||
   regexMatch("log.eventDataCommandLine", "(?i)\\\\.(vbs|vbe|js|jse|wsf|wsh)\\\\s+(//B|//nologo)")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)(CreateObject|WScript\\\\.Shell|Scripting\\\\.FileSystemObject|ADODB\\\\.Stream)") ||
   regexMatch("log.eventDataCommandLine", "(?i)(CreateObject|WScript\\\\.Shell|Scripting\\\\.FileSystemObject|ADODB\\\\.Stream)"))
)
', '2026-03-02 13:27:06.869053', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1045, 'Windows Remote Management (WinRM) Abuse', 3, 3, 2, 'Lateral Movement', 'T1021.006 - Remote Services: Windows Remote Management', e'Detects potential abuse of Windows Remote Management (WinRM) for lateral movement. Monitors for successful logon events (4624) with network logon type 3 combined with privilege escalation (4672) and WinRM-related process activity, indicating remote command execution via WinRM.

Next Steps:
1. Investigate the source IP address and verify if it\'s an authorized administrative workstation
2. Review the target user account for any signs of compromise or unusual privilege usage
3. Examine recent PowerShell execution logs and command history on the target system
4. Check for concurrent suspicious activities from the same source IP across other systems
5. Verify if the WinRM connection aligns with scheduled maintenance or authorized administrative tasks
6. Review network traffic patterns between source and target systems for data exfiltration indicators
7. Validate the legitimacy of any processes spawned through the WinRM session
8. Consider implementing additional monitoring for WinRM usage if this represents unexpected activity
', '["https://jpcertcc.github.io/ToolAnalysisResultSheet/details/WinRM.htm","https://attack.mitre.org/techniques/T1021/006/"]', 'equals("log.eventCode", "4624") && equals("log.eventDataLogonType", "3") && exists("log.eventDataProcessName") && (contains("log.eventDataProcessName", "wsmprovhost.exe") || contains("log.eventDataProcessName", "winrshost.exe") || contains("log.eventDataProcessName", "powershell.exe"))', '2026-03-02 13:27:08.173569', true, true, 'origin', null, '[]', '["lastEvent.target.user","adversary.host","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1046, 'Windows Defender Tampering Detection', 2, 3, 3, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects attempts to disable or tamper with Windows Defender components through registry modifications, service stops, or exclusion additions. This includes registry changes to disable antivirus features, stopping Windows Defender services, or using PowerShell commands to modify protection settings.

Next Steps:
1. Investigate the user account and process responsible for the tampering attempt
2. Check if this is part of authorized maintenance or if it\'s malicious activity
3. Review recent logon events and process execution history on the affected host
4. Verify Windows Defender configuration and ensure it\'s properly restored
5. Scan the system for malware that may have disabled protection
6. Check for other security tool tampering attempts across the environment
', '["https://attack.mitre.org/techniques/T1562/001/","https://docs.microsoft.com/en-us/windows/security/threat-protection/windows-defender-antivirus/troubleshoot-windows-defender-antivirus"]', e'(equals("log.eventId", "4657") &&
 contains("log.eventDataObjectName", "\\\\Windows Defender\\\\") &&
 (contains("log.eventDataObjectValueName", "DisableAntiSpyware") ||
  contains("log.eventDataObjectValueName", "DisableAntiVirus") ||
  contains("log.eventDataObjectValueName", "DisableRealtimeMonitoring"))) ||
(equals("log.eventId", "7040") &&
 contains("log.eventDataServiceName", "WinDefend") &&
 equals("log.eventDataNewState", "disabled")) ||
(equals("log.eventId", "4688") &&
 contains("log.eventDataCommandLine", "Set-MpPreference") &&
 (contains("log.eventDataCommandLine", "DisableRealtimeMonitoring") ||
  contains("log.eventDataCommandLine", "DisableIOAVProtection") ||
  contains("log.eventDataCommandLine", "DisableBehaviorMonitoring")))
', '2026-03-02 13:27:09.488351', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1047, 'Vulnerable Driver Loading Detection (BYOVD)', 3, 3, 3, 'Privilege Escalation', 'T1068 - Exploitation for Privilege Escalation', e'Detects Bring Your Own Vulnerable Driver (BYOVD) attacks where adversaries load known vulnerable
kernel drivers to disable security tools, escalate privileges, or gain kernel-level access.
This technique is increasingly used by ransomware groups and APTs to bypass endpoint protection.
The rule monitors for service installations loading known vulnerable drivers and suspicious
driver-related command patterns.

Next Steps:
1. Identify the loaded driver and check against known vulnerable driver lists (loldrivers.io)
2. Examine the service installation that loaded the driver
3. Check if endpoint security tools were disabled after driver loading
4. Review the process that installed the vulnerable driver
5. Remove the vulnerable driver and restore security tools
6. Block the driver hash via WDAC or application control policies
7. Check for privilege escalation or security tool tampering after driver load
8. Search for the same driver across other endpoints
', '["https://attack.mitre.org/techniques/T1068/","https://www.loldrivers.io/","https://github.com/magicsword-io/LOLDrivers"]', e'(
  equals("log.eventCode", "4697") &&
  equals("log.channel", "Security") &&
  (
    regexMatch("log.eventDataServiceFileName", "(?i)(dbutil_2_3|dbutildrv2|rtcore64|gdrv|asio|ene\\\\.sys|procexp|viragt64|aswarpot|hw64|cpuz|gmer64|mhyprot2|kdmapper|iqvw64e|echo_driver|winio|amifldrv64|elby|vboxdrv|gpu_temp|nicm|phymemx64|micio64|directio64|blacklotus|irec|zemana)") ||
    regexMatch("log.eventDataServiceFileName", "(?i)\\\\.sys\\\\s*$")
  )
) ||
(
  (equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
  (
    regexMatch("log.eventDataCommandLine", "(?i)(sc\\\\.exe|sc)\\\\s+(create|start)\\\\s+.*type=\\\\s*kernel") ||
    regexMatch("log.eventDataCommandLine", "(?i)(sc\\\\.exe|sc)\\\\s+(create|start)\\\\s+.*type=\\\\s*kernel") ||
    regexMatch("log.eventDataCommandLine", "(?i)(sc\\\\.exe|sc)\\\\s+create\\\\s+.*binPath=.*\\\\.sys") ||
    regexMatch("log.eventDataCommandLine", "(?i)(sc\\\\.exe|sc)\\\\s+create\\\\s+.*binPath=.*\\\\.sys")
  )
)
', '2026-03-02 13:27:10.752364', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1048, 'Volume Shadow Copy Deletion', 2, 3, 3, 'Impact', 'T1490 - Inhibit System Recovery', e'Detects deletion of Volume Shadow Copies which is commonly performed by ransomware to prevent recovery of encrypted files. This rule identifies various methods attackers use to delete shadow copies including vssadmin, wmic, PowerShell, wbadmin, and bcdedit commands.

Next Steps:
1. Immediately investigate the affected host for signs of ransomware activity
2. Check for recent file encryption or unusual file modifications
3. Review process execution timeline around the shadow copy deletion
4. Verify if this was legitimate administrative activity or malicious
5. Examine network connections and lateral movement indicators
6. Check backup integrity and consider isolating the system if ransomware is confirmed
7. Review user accounts and privileges involved in the activity
', '["https://attack.mitre.org/techniques/T1490/","https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/vssadmin"]', e'equals("log.eventId", "4688") &&
((endsWith("log.eventDataProcessName", "vssadmin.exe") &&
  (contains("log.eventDataCommandLine", "delete shadows") ||
   (contains("log.eventDataCommandLine", "resize shadowstorage") &&
    contains("log.eventDataCommandLine", "maxsize=")))) ||
 (endsWith("log.eventDataProcessName", "wmic.exe") &&
  contains("log.eventDataCommandLine", "shadowcopy") &&
  contains("log.eventDataCommandLine", "delete")) ||
 (endsWith("log.eventDataProcessName", "powershell.exe") &&
  (contains("log.eventDataCommandLine", "Get-WmiObject") ||
   contains("log.eventDataCommandLine", "gwmi")) &&
  contains("log.eventDataCommandLine", "Win32_ShadowCopy") &&
  contains("log.eventDataCommandLine", "Delete()")) ||
 (endsWith("log.eventDataProcessName", "wbadmin.exe") &&
  contains("log.eventDataCommandLine", "delete") &&
  (contains("log.eventDataCommandLine", "catalog") ||
   contains("log.eventDataCommandLine", "backup"))) ||
 (endsWith("log.eventDataProcessName", "bcdedit.exe") &&
  contains("log.eventDataCommandLine", "recoveryenabled") &&
  contains("log.eventDataCommandLine", "no")))
', '2026-03-02 13:27:12.068201', true, true, 'origin', null, '[]', '["lastEvent.log.eventDataProcessName","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1049, 'UAC Bypass Attempt Detection', 3, 3, 2, 'Privilege Escalation, Defense Evasion', 'T1548.002 - Abuse Elevation Control Mechanism: Bypass User Account Control', e'Detects potential UAC bypass attempts by monitoring for processes with elevated privileges that were not launched through the proper UAC consent mechanism. This rule identifies processes running with elevated token types (TokenElevationTypeVirtual) without proper UAC consent flow, which may indicate an attempt to bypass User Account Control security measures.

Next Steps:
1. Investigate the process that triggered the alert - examine the process name, command line arguments, and parent process
2. Check if the process is known legitimate software that should have elevated privileges
3. Verify the user account that launched the process and their normal privilege level
4. Look for other suspicious activities on the same host around the same time
5. Check for known UAC bypass techniques such as DLL hijacking, registry manipulation, or abuse of auto-elevated processes
6. Examine the process execution chain to identify the attack vector
7. Consider isolating the affected system if malicious activity is confirmed
8. Review security policies and UAC settings to prevent similar bypasses
', '["https://attack.mitre.org/techniques/T1548/002/","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventid=4688"]', e'equals("log.eventId", "4688") &&
equals("log.eventDataTokenElevationType", "2") &&
!contains("log.eventDataParentProcessName", "consent.exe") &&
!contains("log.eventDataProcessName", "TrustedInstaller.exe") &&
!contains("log.eventDataSubjectUserName", "SYSTEM") &&
(contains("log.eventDataParentProcessName", "fodhelper.exe") ||
 contains("log.eventDataParentProcessName", "computerdefaults.exe") ||
 contains("log.eventDataParentProcessName", "sdclt.exe") ||
 contains("log.eventDataParentProcessName", "slui.exe") ||
 contains("log.eventDataParentProcessName", "eventvwr.exe") ||
 contains("log.eventDataParentProcessName", "cmstp.exe") ||
 contains("log.eventDataParentProcessName", "msconfig.exe") ||
 contains("log.eventDataParentProcessName", "dccw.exe") ||
 (contains("log.eventDataProcessName", "cmd.exe") || contains("log.eventDataProcessName", "powershell.exe") || contains("log.eventDataProcessName", "pwsh.exe")))
', '2026-03-02 13:27:13.329683', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1050, 'Tunneling Service C2 Detection', 3, 2, 2, 'Command and Control', 'T1572 - Protocol Tunneling', e'Detects execution of tunneling services (ngrok, cloudflared, devtunnels, localtonet, bore, chisel)
that create reverse tunnels for command and control or to expose internal services to the internet.
These tools are increasingly abused by threat actors for persistent C2 channels that bypass
network security controls because traffic appears as legitimate HTTPS connections.

Next Steps:
1. Identify the tunneling tool being used and its configuration
2. Determine what internal service or port is being exposed
3. Check if this is authorized usage (dev/test) or malicious
4. Review the tunnel endpoint for C2 activity
5. Block the tunneling tool and its associated domains at the firewall
6. Terminate the tunnel process and remove the tool
7. Check for data exfiltration through the tunnel
8. Hunt for the same tunneling tool across other endpoints
', '["https://attack.mitre.org/techniques/T1572/","https://www.microsoft.com/en-us/security/blog/2023/05/24/volt-typhoon-targets-us-critical-infrastructure-with-living-off-the-land-techniques/","https://ngrok.com/abuse"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  regexMatch("log.eventDataNewProcessName", "(?i)(ngrok|cloudflared|devtunnels|localtonet|bore|chisel|frpc|rathole)\\\\.exe$") ||
  regexMatch("log.eventDataProcessName", "(?i)(ngrok|cloudflared|devtunnels|localtonet|bore|chisel|frpc|rathole)\\\\.exe$") ||
  regexMatch("log.eventDataCommandLine", "(?i)ngrok\\\\s+(http|tcp|tls|start)") ||
  regexMatch("log.eventDataCommandLine", "(?i)ngrok\\\\s+(http|tcp|tls|start)") ||
  regexMatch("log.eventDataCommandLine", "(?i)cloudflared\\\\s+tunnel\\\\s+(run|--url)") ||
  regexMatch("log.eventDataCommandLine", "(?i)cloudflared\\\\s+tunnel\\\\s+(run|--url)") ||
  regexMatch("log.eventDataCommandLine", "(?i)devtunnels\\\\s+(host|create|port)") ||
  regexMatch("log.eventDataCommandLine", "(?i)devtunnels\\\\s+(host|create|port)") ||
  regexMatch("log.eventDataCommandLine", "(?i)chisel\\\\s+(client|server)") ||
  regexMatch("log.eventDataCommandLine", "(?i)chisel\\\\s+(client|server)") ||
  regexMatch("log.eventDataCommandLine", "(?i)frpc\\\\.exe.*-c") ||
  regexMatch("log.eventDataCommandLine", "(?i)frpc\\\\.exe.*-c")
)
', '2026-03-02 13:27:14.689098', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1051, 'File Timestomping Detection', 1, 3, 1, 'Defense Evasion', 'T1070.006 - Indicator Removal: Timestomp', e'Detects timestomping activities where attackers modify file creation or modification timestamps
to blend malicious files with legitimate system files. Common tools include PowerShell
Set-ItemProperty, timestomp.exe, NirCmd, and direct API calls to SetFileTime. This technique
is used to evade forensic timeline analysis and make malicious files appear older.

Next Steps:
1. Identify which files had their timestamps modified
2. Check if the modified timestamps match legitimate system file dates (common pattern)
3. Analyze the files with modified timestamps for malicious content
4. Review the process and user that performed the timestamp modification
5. Investigate what the attacker was trying to hide with timestomping
6. Use $MFT analysis to detect original timestamps vs modified ones
7. Search for additional anti-forensic techniques on the same host
', '["https://attack.mitre.org/techniques/T1070/006/","https://www.inversecos.com/2022/04/defence-evasion-technique-timestomping.html","https://andreafortuna.org/2017/10/06/timestomping-what-it-is-and-how-to-detect-it/"]', e'(
  equals("log.eventCode", "2") &&
  equals("log.providerName", "Microsoft-Windows-Sysmon") &&
  exists("log.eventDataPreviousCreationUtcTime") &&
  exists("log.eventDataCreationUtcTime") &&
  !regexMatch("log.eventDataImage", "(?i)(setup|install|update|patch|msiexec|TiWorker|TrustedInstaller|svchost)\\\\.exe$")
) ||
(
  (equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
  (
    regexMatch("log.eventDataCommandLine", "(?i)Set-ItemProperty.*CreationTime|Set-ItemProperty.*LastWriteTime|Set-ItemProperty.*LastAccessTime") ||
    regexMatch("log.eventDataCommandLine", "(?i)Set-ItemProperty.*CreationTime|Set-ItemProperty.*LastWriteTime|Set-ItemProperty.*LastAccessTime") ||
    regexMatch("log.eventDataCommandLine", "(?i)timestomp|SetFileTime|NirCmd.*setfiletime") ||
    regexMatch("log.eventDataCommandLine", "(?i)timestomp|SetFileTime|NirCmd.*setfiletime") ||
    regexMatch("log.eventDataCommandLine", "(?i)\\\\[IO\\\\.File\\\\]::SetCreationTime|\\\\[IO\\\\.File\\\\]::SetLastWriteTime") ||
    regexMatch("log.eventDataCommandLine", "(?i)\\\\[IO\\\\.File\\\\]::SetCreationTime|\\\\[IO\\\\.File\\\\]::SetLastWriteTime")
  )
)
', '2026-03-02 13:27:15.999710', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1052, 'Sysmon Service Tampering Detection', 3, 3, 3, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects attempts to tamper with the Sysmon monitoring service including unloading the Sysmon
driver, stopping or deleting the Sysmon service, modifying Sysmon configuration, or renaming
the Sysmon binary. Attackers target Sysmon specifically because it provides rich endpoint
telemetry that enables threat detection. Disabling Sysmon is a critical defense evasion step.

Next Steps:
1. Immediately investigate the process and user that tampered with Sysmon
2. Restore Sysmon service and verify it is running correctly
3. Check if the Sysmon driver is still loaded in the kernel
4. Review what malicious activity occurred during the Sysmon outage
5. Re-apply the Sysmon configuration from a known good backup
6. Investigate the full attack chain that led to Sysmon tampering
7. Consider implementing tamper protection for Sysmon
', '["https://attack.mitre.org/techniques/T1562/001/","https://docs.microsoft.com/en-us/sysinternals/downloads/sysmon","https://undev.ninja/sysmon-evasion-techniques/"]', e'(
  (equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
  (
    (regexMatch("log.eventDataCommandLine", "(?i)sysmon.*-u") ||
     regexMatch("log.eventDataCommandLine", "(?i)sysmon.*-u")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)sc(.exe)?\\\\s+(stop|delete|config)\\\\s+sysmon") ||
     regexMatch("log.eventDataCommandLine", "(?i)sc(.exe)?\\\\s+(stop|delete|config)\\\\s+sysmon")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)net(.exe)?\\\\s+stop\\\\s+sysmon") ||
     regexMatch("log.eventDataCommandLine", "(?i)net(.exe)?\\\\s+stop\\\\s+sysmon")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)fltMC(.exe)?\\\\s+unload\\\\s+sysmon") ||
     regexMatch("log.eventDataCommandLine", "(?i)fltMC(.exe)?\\\\s+unload\\\\s+sysmon")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)taskkill.*/f.*/im\\\\s+sysmon") ||
     regexMatch("log.eventDataCommandLine", "(?i)taskkill.*/f.*/im\\\\s+sysmon"))
  )
) ||
(
  equals("log.eventCode", "7045") &&
  regexMatch("log.eventDataServiceName", "(?i)sysmon") &&
  !equals("log.eventDataServiceType", "kernel mode driver")
)
', '2026-03-02 13:27:17.398400', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1053, 'Suspicious Service Installation Detection', 3, 3, 2, 'Persistence', 'T1543.003 - Create or Modify System Process: Windows Service', e'Detects installation of suspicious Windows services matching patterns from Cobalt Strike, Metasploit,
Impacket PSExec, and Meterpreter payloads. Attackers commonly install malicious services for persistence,
privilege escalation (getsystem), and lateral movement. The rule monitors Event ID 4697 for service
installations with characteristic names, binary paths, or command-line patterns used by offensive frameworks.

Next Steps:
1. Examine the service binary path for malicious executables or scripts
2. Check if the service name matches known offensive tool patterns
3. Verify the installing user account and their authorization level
4. Review the service configuration for suspicious startup types
5. Stop and disable the suspicious service immediately
6. Analyze the service binary in a sandbox environment
7. Search for lateral movement indicators from the source host
8. Check for other persistence mechanisms installed by the same actor
', '["https://attack.mitre.org/techniques/T1543/003/","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4697","https://www.sans.org/blog/red-team-tactics-hiding-windows-services/"]', e'equals("log.eventCode", "4697") &&
equals("log.channel", "Security") &&
(
  regexMatch("log.eventDataServiceFileName", "(?i)(cmd\\\\.exe\\\\s*/c|powershell|pwsh|%COMSPEC%)") ||
  regexMatch("log.eventDataServiceFileName", "(?i)(\\\\\\\\ADMIN\\\\$|\\\\\\\\C\\\\$|127\\\\.0\\\\.0\\\\.1)") ||
  regexMatch("log.eventDataServiceFileName", "(?i)(rundll32|msiexec|regsvr32|mshta|certutil|bitsadmin)") ||
  regexMatch("log.eventDataServiceName", "(?i)^[a-zA-Z]{4,8}$") ||
  regexMatch("log.eventDataServiceFileName", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\|\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Users\\\\\\\\Public\\\\\\\\)") ||
  regexMatch("log.eventDataServiceFileName", "(?i)\\\\.exe\\\\s+[a-zA-Z0-9+/=]{50,}") ||
  regexMatch("log.eventDataServiceName", "(?i)(BTOBTO|meterpreter|msf|payload|beacon|cobalt)") ||
  regexMatch("log.eventDataServiceFileName", "(?i)(comsvcs\\\\.dll|MiniDump|procdump|lsass)")
)
', '2026-03-02 13:27:18.887636', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1054, 'Startup Folder Persistence Detection', 2, 3, 2, 'Persistence', 'T1547.001 - Boot or Logon Autostart Execution: Registry Run Keys / Startup Folder', e'Detects file creation or modification in Windows Startup folders, which execute programs
automatically on user login. Attackers drop malicious scripts, executables, or shortcut files
into Startup folders for persistence. The rule monitors both per-user and system-wide Startup
directories for suspicious file types.

Next Steps:
1. Identify the file dropped into the Startup folder and analyze its contents
2. Check the parent process that created the file for suspicious origins
3. Verify if the file is a legitimate application shortcut or a malicious payload
4. Remove the suspicious file from the Startup folder
5. Analyze the executable or script for malicious behavior in a sandbox
6. Search for additional persistence mechanisms on the same host
7. Investigate the initial access vector
', '["https://attack.mitre.org/techniques/T1547/001/","https://docs.microsoft.com/en-us/windows/win32/shell/manage-program-startup","https://www.cybereason.com/blog/persistence-techniques-that-persist"]', e'(
  equals("log.eventCode", "11") &&
  equals("log.providerName", "Microsoft-Windows-Sysmon") &&
  (
    regexMatch("log.eventDataTargetFilename", "(?i)\\\\\\\\Start Menu\\\\\\\\Programs\\\\\\\\Startup\\\\\\\\") ||
    regexMatch("log.eventDataTargetFilename", "(?i)\\\\\\\\ProgramData\\\\\\\\Microsoft\\\\\\\\Windows\\\\\\\\Start Menu\\\\\\\\Programs\\\\\\\\StartUp\\\\\\\\")
  ) &&
  regexMatch("log.eventDataTargetFilename", "(?i)\\\\.(exe|dll|bat|cmd|vbs|vbe|js|jse|wsf|wsh|ps1|hta|lnk|scr|pif)$")
) ||
(
  (equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
  (
    (regexMatch("log.eventDataCommandLine", "(?i)copy.*\\\\\\\\Start Menu\\\\\\\\Programs\\\\\\\\Startup") ||
     regexMatch("log.eventDataCommandLine", "(?i)copy.*\\\\\\\\Start Menu\\\\\\\\Programs\\\\\\\\Startup")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)move.*\\\\\\\\Start Menu\\\\\\\\Programs\\\\\\\\Startup") ||
     regexMatch("log.eventDataCommandLine", "(?i)move.*\\\\\\\\Start Menu\\\\\\\\Programs\\\\\\\\Startup"))
  )
)
', '2026-03-02 13:27:20.019454', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1055, 'SMBv1 Usage Detection', 3, 2, 2, 'Lateral Movement', 'T1210 - Exploitation of Remote Services', e'Detects usage of the deprecated and vulnerable SMBv1 protocol which could be exploited for lateral movement or ransomware propagation. SMBv1 is susceptible to numerous security vulnerabilities including EternalBlue and should be disabled in favor of SMBv2/SMBv3.

Next Steps:
1. Immediately investigate the source system using SMBv1 and identify which service or application is still dependent on this protocol
2. Review network traffic logs to determine if this is internal communication or external access attempts
3. Check for any signs of exploitation attempts or successful compromises on the affected system
4. Identify all systems in the environment that may still have SMBv1 enabled
5. Plan migration to SMBv2/SMBv3 and disable SMBv1 on all systems where possible
6. Monitor for any lateral movement patterns that may indicate ongoing compromise
7. Consider implementing network segmentation to limit exposure if SMBv1 cannot be immediately disabled
', '["https://learn.microsoft.com/en-us/windows-server/storage/file-server/troubleshoot/detect-enable-and-disable-smbv1-v2-v3","https://attack.mitre.org/techniques/T1210/"]', 'equals("log.eventId", "3000") && equals("log.providerName", "Microsoft-Windows-SMBServer") && contains("log.message", "SMB1")', '2026-03-02 13:27:21.193572', true, true, 'origin', null, '[]', '["adversary.host","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1056, 'Sliver C2 Framework Detection', 3, 3, 2, 'Execution', 'T1059 - Command and Scripting Interpreter', e'Detects Sliver C2 framework execution patterns and artifacts. Sliver is an increasingly popular
open-source C2 framework replacing Cobalt Strike in many threat actor toolkits. The rule monitors
for characteristic Sliver implant patterns including specific PowerShell stager patterns,
named pipe names, and executable naming conventions.

Next Steps:
1. Identify the Sliver implant type (HTTP, HTTPS, mTLS, DNS, WireGuard)
2. Check for network beaconing to C2 infrastructure
3. Review the process tree for injection and post-exploitation activity
4. Memory scan affected processes for Sliver implant shellcode
5. Isolate the compromised host immediately
6. Block identified C2 domains and IPs at the network perimeter
7. Hunt for Sliver artifacts across all endpoints
8. Review initial access vector (phishing, exploitation, etc.)
', '["https://github.com/BishopFox/sliver","https://www.microsoft.com/en-us/security/blog/2022/08/24/looking-for-the-sliver-lining-hunting-for-emerging-command-and-control-frameworks/","https://www.cybereason.com/blog/sliver-c2-leveraged-by-many-threat-actors"]', e'(
  (equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
  (
    regexMatch("log.eventDataCommandLine", "(?i)(sliver|sliverpb|silver-implant)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(sliver|sliverpb|silver-implant)") ||
    regexMatch("log.eventDataCommandLine", "(?i)\\\\\\\\\\\\\\\\.\\\\\\\\pipe\\\\\\\\(sliverpb|sliver-)") ||
    regexMatch("log.eventDataCommandLine", "(?i)\\\\\\\\\\\\\\\\.\\\\\\\\pipe\\\\\\\\(sliverpb|sliver-)") ||
    regexMatch("log.eventDataNewProcessName", "(?i)(DECISIVE_|MANAGING_|IMPRESSED_|LITERARY_|ENORMOUS_)") ||
    regexMatch("log.eventDataCommandLine", "(?i)\\\\$s=New-Object\\\\s+IO\\\\.MemoryStream.*FromBase64String.*IO\\\\.Compression\\\\.GzipStream.*IO\\\\.Compression\\\\.CompressionMode.*ReadToEnd") ||
    regexMatch("log.eventDataCommandLine", "(?i)\\\\$s=New-Object\\\\s+IO\\\\.MemoryStream.*FromBase64String.*IO\\\\.Compression\\\\.GzipStream")
  )
) ||
(
  (equals("log.eventCode", "4104") || equals("log.eventId", 4104)) &&
  equals("log.providerName", "Microsoft-Windows-PowerShell") &&
  (
    regexMatch("log.eventDataScriptBlockText", "(?i)\\\\$s=New-Object\\\\s+IO\\\\.MemoryStream.*IO\\\\.Compression\\\\.GzipStream.*IO\\\\.Compression\\\\.CompressionMode.*ReadToEnd") ||
    contains("log.eventDataScriptBlockText", "sliverpb") ||
    regexMatch("log.eventDataScriptBlockText", "(?i)New-Object\\\\s+System\\\\.Net\\\\.Sockets\\\\.TCPClient.*IO\\\\.StreamReader.*IO\\\\.StreamWriter.*GetStream")
  )
)
', '2026-03-02 13:27:22.423356', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1057, 'SID History Injection Attempt', 3, 3, 1, 'Defense Evasion, Privilege Escalation', 'T1134.005 - Access Token Manipulation: SID-History Injection', e'Detects attempts to add SID History to an account, which can be used for privilege escalation. SID History injection allows attackers to inherit permissions from privileged accounts without being members of privileged groups. Both successful (4765) and failed (4766) attempts are monitored.

Next Steps:
1. Immediately investigate the target user account and verify if SID History modification was legitimate
2. Check if the user performing the action has proper administrative privileges for this operation
3. Review the source SID being added to understand what permissions are being inherited
4. Examine recent authentication logs for the target account to identify potential unauthorized access
5. Verify Active Directory configuration and check for signs of domain controller compromise
6. Consider resetting the target account password and removing unauthorized SID History entries
7. Review domain administrator accounts and privileged group memberships for anomalies
', '["https://attack.mitre.org/techniques/T1134/005/","https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4765","https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4766","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventid=4765"]', e'oneOf("log.eventId", ["4765", "4766"]) &&
equals("log.channel", "Security")
', '2026-03-02 13:27:23.721333', true, true, 'origin', null, '[]', '["adversary.user","target.host","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1058, 'Shadow Credentials Attack Detection', 3, 3, 1, 'Credential Access', 'T1556 - Modify Authentication Process', e'Detects Shadow Credentials attacks where adversaries modify the msDS-KeyCredentialLink attribute
on Active Directory objects to add their own key credentials for passwordless authentication.
This enables attackers to authenticate as the target account without knowing the password,
effectively taking over the account. The rule monitors Event ID 5136 (directory service object
modification) and Event ID 4662 (object access) for msDS-KeyCredentialLink changes.

Next Steps:
1. Identify the account whose msDS-KeyCredentialLink was modified
2. Check if the modification was done by an authorized admin or service
3. Examine the added key credential for malicious certificates
4. Remove unauthorized key credentials from the affected account
5. Reset the password for the affected account
6. Review who has write access to the msDS-KeyCredentialLink attribute
7. Audit all accounts for unauthorized key credential additions
8. Investigate for related tools like Whisker or pyWhisker
', '["https://attack.mitre.org/techniques/T1556/","https://posts.specterops.io/shadow-credentials-abusing-key-trust-account-mapping-for-takeover-8ee1a53566ab","https://github.com/eladshamir/Whisker"]', e'(
  equals("log.eventCode", "5136") &&
  equals("log.channel", "Security") &&
  contains("log.eventDataAttributeLDAPDisplayName", "msDS-KeyCredentialLink")
) ||
(
  equals("log.eventCode", "4662") &&
  equals("log.channel", "Security") &&
  contains("log.eventDataProperties", "msDS-KeyCredentialLink") &&
  !regexMatch("log.eventDataSubjectUserName", "(?i)\\\\$$")
)
', '2026-03-02 13:27:24.871464', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1059, 'Suspicious Scheduled Task Persistence', 2, 3, 2, 'Persistence', 'T1053.005 - Scheduled Task/Job: Scheduled Task', e'Detects creation of scheduled tasks from suspicious locations or executing suspicious binaries,
which is one of the most common persistence mechanisms used by attackers. The rule monitors
Event ID 4698 (scheduled task creation) for tasks that reference suspicious paths such as
Temp, Downloads, AppData, or ProgramData directories, or execute known LOLBINs like
PowerShell, certutil, rundll32, mshta, regsvr32, or cmd.exe with suspicious arguments.

Next Steps:
1. Examine the full TaskContent XML to understand what the scheduled task executes
2. Identify the user account that created the task and verify authorization
3. Check the task execution path for malicious payloads or scripts
4. Review the trigger schedule for persistence timing patterns
5. Remove the suspicious scheduled task immediately
6. Search for additional persistence mechanisms on the same host
7. Investigate the initial access vector that led to task creation
', '["https://attack.mitre.org/techniques/T1053/005/","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4698","https://posts.specterops.io/abstracting-scheduled-task-creation-6d22febc27e7"]', e'equals("log.eventCode", "4698") &&
equals("log.channel", "Security") &&
(
  regexMatch("log.eventDataTaskContent", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\|\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Public\\\\\\\\|\\\\\\\\Users\\\\\\\\Public\\\\\\\\)") ||
  regexMatch("log.eventDataTaskContent", "(?i)(powershell|pwsh|cmd\\\\.exe|certutil|mshta|rundll32|regsvr32|cscript|wscript|msiexec|bitsadmin)") ||
  regexMatch("log.eventDataTaskContent", "(?i)(-enc\\\\s|-encodedcommand|-nop|-w\\\\s+hidden|downloadstring|invoke-expression|iex)") ||
  regexMatch("log.eventDataTaskContent", "(?i)(http://|https://|ftp://|\\\\\\\\\\\\\\\\[0-9]+\\\\.[0-9]+\\\\.[0-9]+\\\\.[0-9]+\\\\\\\\)")
) &&
!regexMatch("log.eventDataSubjectUserName", "(?i)^(SYSTEM|LOCAL SERVICE|NETWORK SERVICE)$")
', '2026-03-02 13:27:26.262107', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1060, 'SAM Database Access Attempt', 3, 3, 1, 'Credential Access', 'T1003.002 - OS Credential Dumping: Security Account Manager', e'Detects attempts to access the Security Account Manager (SAM) database, which contains local user account hashes. This activity may indicate credential dumping attempts by attackers trying to extract password hashes for offline cracking or lateral movement.

Next Steps:
1. Immediately investigate the user account and process that accessed the SAM database
2. Check for any unusual processes running on the affected system
3. Review recent logon events and privilege escalation activities
4. Examine network connections from the affected host for lateral movement
5. Consider isolating the affected system if malicious activity is confirmed
6. Review security policies for SAM database access permissions
7. Check for presence of credential dumping tools or suspicious files
', '["https://attack.mitre.org/techniques/T1003/002/","https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-10/security/threat-protection/auditing/event-4661"]', e'equals("log.eventCode", "4663") &&
equals("log.channel", "Security") &&
(
  endsWith("log.eventDataObjectName", "\\\\SAM") ||
  endsWith("log.eventDataObjectName", "\\\\SECURITY") ||
  endsWith("log.eventDataObjectName", "\\\\SYSTEM")
) &&
oneOf("log.eventDataAccessMask", ["0x20019", "0x1f01ff", "0x40", "0x20", "0x1"])
', '2026-03-02 13:27:27.362742', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1061, 'Rundll32 Suspicious Abuse Detection', 2, 3, 2, 'Defense Evasion', 'T1218.011 - System Binary Proxy Execution: Rundll32', e'Detects suspicious rundll32.exe usage patterns including loading DLLs from non-system paths,
comsvcs.dll MiniDump for LSASS dumping, JavaScript/VBScript execution, and loading from Temp
or Downloads directories. Rundll32 is a signed Microsoft binary frequently abused for defense
evasion and code execution.

Next Steps:
1. Examine the DLL path and export function being called
2. Verify the DLL is not loading from a suspicious location (Temp, Downloads, AppData)
3. Check the parent process for delivery mechanism indicators
4. Analyze the loaded DLL in a sandbox environment
5. Review network connections made by the rundll32 process
6. Check for process injection from the rundll32 process
7. Search for the same DLL hash across other endpoints
', '["https://attack.mitre.org/techniques/T1218/011/","https://lolbas-project.github.io/lolbas/Binaries/Rundll32/","https://redcanary.com/threat-detection-report/techniques/rundll32/"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  regexMatch("log.eventDataNewProcessName", "(?i)rundll32\\\\.exe$") ||
  regexMatch("log.eventDataProcessName", "(?i)rundll32\\\\.exe$")
) &&
(
  regexMatch("log.eventDataCommandLine", "(?i)comsvcs\\\\.dll.*(MiniDump|#24)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(javascript:|vbscript:)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\|\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Users\\\\\\\\Public\\\\\\\\).*\\\\.dll") ||
  regexMatch("log.eventDataCommandLine", "(?i)shell32\\\\.dll.*ShellExec_RunDLL") ||
  regexMatch("log.eventDataCommandLine", "(?i)advpack\\\\.dll.*RegisterOCX") ||
  regexMatch("log.eventDataCommandLine", "(?i)url\\\\.dll.*FileProtocolHandler") ||
  regexMatch("log.eventDataCommandLine", "(?i)zipfldr\\\\.dll.*RouteTheCall") ||
  regexMatch("log.eventDataCommandLine", "(?i)comsvcs\\\\.dll.*(MiniDump|#24)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(javascript:|vbscript:)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\).*\\\\.dll")
)
', '2026-03-02 13:27:28.578959', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1062, 'Registry Run Key Persistence Detection', 2, 3, 2, 'Persistence', 'T1547.001 - Boot or Logon Autostart Execution: Registry Run Keys / Startup Folder', e'Detects modifications to Windows Registry Run and RunOnce keys, which are the most fundamental
Windows persistence mechanism. Attackers add entries to these keys to execute malicious code
every time a user logs on or the system starts. The rule monitors Event ID 4657 (registry value
modifications) for changes to Run/RunOnce keys with suspicious values pointing to unusual
executable paths, scripts, or LOLBINs.

Next Steps:
1. Examine the registry value data to identify the persistence payload
2. Verify if the modification was made by a legitimate software installer
3. Check the modifying process and user account for compromise indicators
4. Remove the malicious registry entry immediately
5. Analyze the payload binary or script referenced in the registry value
6. Search for additional persistence mechanisms on the same host
7. Review the timeline of events leading to the registry modification
', '["https://attack.mitre.org/techniques/T1547/001/","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4657","https://docs.microsoft.com/en-us/windows/win32/setupapi/run-and-runonce-registry-keys"]', e'equals("log.eventCode", "4657") &&
equals("log.channel", "Security") &&
regexMatch("log.eventDataObjectName", "(?i)(\\\\\\\\Run\\\\\\\\|\\\\\\\\RunOnce\\\\\\\\|\\\\\\\\RunOnceEx\\\\\\\\|\\\\\\\\RunServices\\\\\\\\|\\\\\\\\RunServicesOnce\\\\\\\\)") &&
(
  regexMatch("log.eventDataObjectValueName", "(?i)(powershell|cmd|mshta|rundll32|regsvr32|certutil|wscript|cscript|bitsadmin)") ||
  regexMatch("log.eventDataNewValue", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\|\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Users\\\\\\\\Public\\\\\\\\)") ||
  regexMatch("log.eventDataNewValue", "(?i)(powershell|cmd\\\\.exe\\\\s*/c|mshta|rundll32|regsvr32|certutil|wscript|cscript)") ||
  regexMatch("log.eventDataNewValue", "(?i)(http://|https://|\\\\\\\\\\\\\\\\[0-9]+\\\\.[0-9]+)") ||
  regexMatch("log.eventDataNewValue", "(?i)(-enc\\\\s|-encodedcommand|-w\\\\s+hidden)")
)
', '2026-03-02 13:27:29.797982', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1063, 'RDP Session Hijacking Detection', 3, 3, 2, 'Lateral Movement', 'T1563.002 - Remote Service Session Hijacking: RDP Hijacking', e'Detects RDP session hijacking using tscon.exe, which allows a user with SYSTEM privileges
to connect to another user\'s RDP session without authentication. This is commonly used
after privilege escalation to gain access to active sessions of other users, potentially
including domain administrators.

Next Steps:
1. Identify which user\'s session was hijacked via tscon
2. Verify the SYSTEM-level access method used (service, scheduled task, etc.)
3. Check what actions were performed in the hijacked session
4. Review if the target session had access to sensitive resources
5. Force logout the compromised session
6. Investigate how the attacker obtained SYSTEM privileges
7. Review all active RDP sessions on the affected server
', '["https://attack.mitre.org/techniques/T1563/002/","https://medium.com/@youraveragetechnoob/rdp-session-hijacking-55ef3f85feaa","https://www.keysight.com/blogs/tech/nwvs/2022/09/21/rdp-session-hijacking"]', e'(equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
(
  (regexMatch("log.eventDataCommandLine", "(?i)tscon(.exe)?\\\\s+\\\\d+\\\\s+/dest:") ||
   regexMatch("log.eventDataCommandLine", "(?i)tscon(.exe)?\\\\s+\\\\d+\\\\s+/dest:")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)query\\\\s+session.*tscon") ||
   regexMatch("log.eventDataCommandLine", "(?i)query\\\\s+session.*tscon")) ||
  (regexMatch("log.eventDataNewProcessName", "(?i)\\\\\\\\tscon\\\\.exe$") &&
   regexMatch("log.eventDataCommandLine", "(?i)\\\\d+\\\\s+/dest:"))
)
', '2026-03-02 13:27:31.156268', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1064, 'PsExec Lateral Movement Detection', 3, 3, 2, 'Lateral Movement', 'T1569.002 - System Services: Service Execution', e'Detects PsExec-style lateral movement by monitoring for the creation of the PSEXESVC service
(Event ID 7045), PsExec named pipe creation, and process execution patterns characteristic
of PsExec and its variants (Impacket PSExec, Metasploit PSExec). PsExec is the most common
tool for lateral movement in enterprise environments.

Next Steps:
1. Identify the source host that initiated the PsExec connection
2. Verify if this is authorized administrative activity
3. Check what command was executed via PsExec on the target host
4. Review the service account credentials used for PsExec authentication
5. Search for evidence of PsExec usage across other hosts in the environment
6. Check for credential theft that may have enabled the lateral movement
7. Block PsExec at the network level if not authorized
', '["https://attack.mitre.org/techniques/T1569/002/","https://jpcertcc.github.io/ToolAnalysisResultSheet/details/PsExec.htm","https://www.sans.org/blog/psexec-and-lateral-movement/"]', e'(
  equals("log.eventCode", "7045") &&
  (
    regexMatch("log.eventDataServiceName", "(?i)^(PSEXESVC|psexec|PAExec|csexec|remcom)") ||
    regexMatch("log.eventDataServiceFileName", "(?i)(PSEXESVC|psexec|PAExec)\\\\.exe") ||
    (regexMatch("log.eventDataServiceFileName", "(?i)%SystemRoot%\\\\\\\\[a-zA-Z]{8}\\\\.exe") &&
     regexMatch("log.eventDataServiceName", "(?i)^[a-zA-Z]{8}$")) ||
    regexMatch("log.eventDataServiceFileName", "(?i)cmd\\\\.exe\\\\s+/c.*\\\\\\\\\\\\\\\\[0-9]+\\\\.[0-9]+\\\\.[0-9]+\\\\.[0-9]+\\\\\\\\")
  )
) ||
(
  equals("log.eventCode", "17") &&
  equals("log.providerName", "Microsoft-Windows-Sysmon") &&
  regexMatch("log.eventDataPipeName", "(?i)\\\\\\\\(psexesvc|paexec|csexec|remcom)")
) ||
(
  equals("log.eventCode", "4688") &&
  regexMatch("log.eventDataNewProcessName", "(?i)\\\\\\\\PSEXESVC\\\\.exe$") &&
  exists("log.eventDataParentProcessName")
)
', '2026-03-02 13:27:32.516639', true, true, 'origin', null, '[]', '["adversary.host","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1065, 'Print Spooler Exploitation Detection', 3, 3, 2, 'Privilege Escalation', 'T1068 - Exploitation for Privilege Escalation', e'Detects potential PrintNightmare exploitation attempts through suspicious print spooler activity including DLL loading and driver installation. This rule identifies suspicious activity related to the Print Spooler service that could indicate CVE-2021-34527 (PrintNightmare) exploitation attempts.

Next Steps:
1. Investigate the affected host for unauthorized driver installations
2. Check for recent DLL files placed in the print spooler drivers directory
3. Review print spooler service logs for unusual activity
4. Examine process execution context and parent processes
5. Verify if print driver installations were authorized
6. Check for lateral movement from the affected system
7. Consider isolating the affected host if exploitation is confirmed
', '["https://msrc.microsoft.com/update-guide/vulnerability/CVE-2021-34527","https://attack.mitre.org/techniques/T1068/"]', '(equals("log.providerName", "Microsoft-Windows-PrintService") && equals("log.eventId", "316") && contains("log.message", "kernelbase.dll")) || (equals("log.eventDataProcessName", "spoolsv.exe") && contains("log.eventDataTargetFilename", "\\spool\\drivers\\x64\\") && endsWith("log.eventDataTargetFilename", ".dll"))', '2026-03-02 13:27:33.993741', true, true, 'origin', null, '[]', '["lastEvent.log.eventDataTargetFilename","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1066, 'PowerShell Empire Detection', 3, 3, 2, 'Execution', 'T1059.001 - Command and Scripting Interpreter: PowerShell', e'Detects potential PowerShell Empire framework usage based on characteristic command patterns, obfuscation techniques, and encoded payloads commonly used by this post-exploitation framework. PowerShell Empire is a post-exploitation framework that uses PowerShell and Python agents to maintain persistence and execute commands on compromised systems.

Next Steps:
1. Immediately isolate the affected host to prevent lateral movement
2. Analyze the complete PowerShell script block content for additional IOCs
3. Check for persistence mechanisms (scheduled tasks, registry entries, services)
4. Review network connections from the host for C2 communication
5. Examine process tree and parent processes that spawned PowerShell
6. Search for additional Empire artifacts across the environment
7. Reset credentials for any accounts used on the compromised system
8. Conduct memory analysis to identify injected code or payloads
9. Review recent user activity and file access patterns
10. Update endpoint detection rules based on specific Empire techniques observed
', '["https://attack.mitre.org/techniques/T1059/001/","https://www.powershellempire.com/","https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_logging"]', e'(equals("log.eventCode", "4104") || equals("log.eventId", 4104)) &&
equals("log.providerName", "Microsoft-Windows-PowerShell") &&
(
  contains("log.eventDataScriptBlockText", "System.Management.Automation.AmsiUtils") ||
  regexMatch("log.eventDataScriptBlockText", "(?i)(empire|invoke-empire|invoke-psempire)") ||
  regexMatch("log.eventDataScriptBlockText", "(?i)\\\\[System\\\\.Convert\\\\]::FromBase64String") ||
  regexMatch("log.eventDataScriptBlockText", "(?i)IEX\\\\s*\\\\(\\\\s*New-Object") ||
  regexMatch("log.eventDataScriptBlockText", "(?i)-enc\\\\s+[A-Za-z0-9+/=]{100,}") ||
  regexMatch("log.eventDataScriptBlockText", "(?i)\\\\$DoIt\\\\s*=\\\\s*@") ||
  regexMatch("log.eventDataScriptBlockText", "(?i)\\\\[System\\\\.Text\\\\.Encoding\\\\]::Unicode\\\\.GetString") ||
  contains("log.eventDataScriptBlockText", "Invoke-Shellcode") ||
  contains("log.eventDataScriptBlockText", "Invoke-ReflectivePEInjection") ||
  contains("log.eventDataScriptBlockText", "Invoke-Mimikatz")
)
', '2026-03-02 13:27:35.265686', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1067, 'Pass-the-Ticket Attack Detection', 3, 3, 2, 'Lateral Movement', 'T1550.003 - Use Alternate Authentication Material: Pass the Ticket', e'Detects Pass-the-Ticket attacks where attackers steal and reuse Kerberos tickets from one
endpoint to authenticate on another. The rule monitors for Kerberos authentication events
(Event ID 4624 LogonType 3 with Kerberos package) combined with anomalous ticket usage
patterns and tools like Rubeus, Mimikatz kerberos::ptt, and Invoke-Mimikatz.

Next Steps:
1. Identify the source of the Kerberos ticket and the target system
2. Check if the ticket was exported using Mimikatz or Rubeus
3. Verify if the original ticket owner\'s host is compromised
4. Review Kerberos TGS/TGT events from the source IP
5. Force Kerberos ticket renewal by resetting the user\'s password
6. Check for additional lateral movement from the authenticated session
7. Implement Kerberos constrained delegation and armoring
', '["https://attack.mitre.org/techniques/T1550/003/","https://adsecurity.org/?p=2011","https://www.sans.org/blog/kerberos-in-the-crosshairs-golden-tickets-silver-tickets-mitm-and-more/"]', e'(
  (equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
  (
    (regexMatch("log.eventDataCommandLine", "(?i)kerberos::ptt") ||
     regexMatch("log.eventDataCommandLine", "(?i)kerberos::ptt")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)Rubeus.*ptt") ||
     regexMatch("log.eventDataCommandLine", "(?i)Rubeus.*ptt")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)Rubeus.*(asktgt|asktgs|renew|s4u|createnetonly)") ||
     regexMatch("log.eventDataCommandLine", "(?i)Rubeus.*(asktgt|asktgs|renew|s4u|createnetonly)")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)Invoke-Mimikatz.*kerberos") ||
     regexMatch("log.eventDataCommandLine", "(?i)Invoke-Mimikatz.*kerberos")) ||
    (regexMatch("log.eventDataCommandLine", "(?i)klist.*purge|klist.*tickets") ||
     regexMatch("log.eventDataCommandLine", "(?i)klist.*purge|klist.*tickets"))
  )
) ||
(
  equals("log.eventCode", "4768") &&
  equals("log.channel", "Security") &&
  !equals("log.eventDataStatus", "0x0") &&
  oneOf("log.eventDataStatus", ["0x1f", "0x20", "0x25"]) &&
  exists("log.eventDataIpAddress")
)
', '2026-03-02 13:27:36.357874', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1068, 'NTLM Authentication Downgrade Attack', 3, 3, 1, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects NTLM authentication downgrade attacks via registry modifications to LMCompatibilityLevel,
NtlmMinClientSec, and NtlmMinServerSec. Attackers modify these registry values to weaken NTLM
authentication security, enabling credential interception, relay attacks, and offline cracking
of captured NTLM hashes. Downgrading to LM or NTLMv1 authentication makes credential theft
significantly easier.

Next Steps:
1. Check the new registry value to determine the downgrade severity
2. LMCompatibilityLevel < 3 enables NTLMv1 which is trivially crackable
3. Identify the process and user that made the registry modification
4. Restore the registry values to enforce NTLMv2 (LMCompatibilityLevel = 5)
5. Review Group Policy for conflicting NTLM settings
6. Check for NTLM relay attacks following the downgrade
7. Audit network traffic for NTLMv1 authentication attempts
8. Consider disabling NTLM entirely where possible
', '["https://attack.mitre.org/techniques/T1562/001/","https://www.praetorian.com/blog/ntlm-relaying-attacks/","https://docs.microsoft.com/en-us/windows/security/threat-protection/security-policy-settings/network-security-lan-manager-authentication-level"]', e'equals("log.eventCode", "4657") &&
equals("log.channel", "Security") &&
(
  regexMatch("log.eventDataObjectName", "(?i)\\\\\\\\CurrentControlSet\\\\\\\\Control\\\\\\\\Lsa\\\\\\\\LMCompatibilityLevel") ||
  regexMatch("log.eventDataObjectName", "(?i)\\\\\\\\CurrentControlSet\\\\\\\\Control\\\\\\\\Lsa\\\\\\\\MSV1_0\\\\\\\\NtlmMinClientSec") ||
  regexMatch("log.eventDataObjectName", "(?i)\\\\\\\\CurrentControlSet\\\\\\\\Control\\\\\\\\Lsa\\\\\\\\MSV1_0\\\\\\\\NtlmMinServerSec") ||
  regexMatch("log.eventDataObjectName", "(?i)\\\\\\\\CurrentControlSet\\\\\\\\Control\\\\\\\\Lsa\\\\\\\\MSV1_0\\\\\\\\RestrictSendingNTLMTraffic") ||
  regexMatch("log.eventDataObjectName", "(?i)\\\\\\\\CurrentControlSet\\\\\\\\Control\\\\\\\\Lsa\\\\\\\\MSV1_0\\\\\\\\AuditReceivingNTLMTraffic") ||
  regexMatch("log.eventDataObjectName", "(?i)\\\\\\\\CurrentControlSet\\\\\\\\Services\\\\\\\\Netlogon\\\\\\\\Parameters\\\\\\\\RequireSignOrSeal")
) &&
!regexMatch("log.eventDataSubjectUserName", "(?i)^(SYSTEM|TrustedInstaller)$")
', '2026-03-02 13:27:37.625530', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1069, 'Network Sniffer and Packet Capture Detection', 3, 1, 0, 'Discovery', 'T1040 - Network Sniffing', e'Detects execution of network sniffing and packet capture tools that could be used to
intercept network traffic, capture credentials, or perform man-in-the-middle attacks.
Monitors for known sniffer binaries, raw socket creation, and promiscuous mode indicators.

Next Steps:
1. Verify if the packet capture tool usage was authorized (IT/security staff)
2. Check which user account executed the tool and validate their role
3. Determine what network interfaces were targeted for capture
4. Review if any sensitive data (credentials, tokens) may have been captured
5. Check for data exfiltration - captured packets being sent to external destinations
6. Look for lateral movement from the host running the sniffer
7. Investigate if the sniffer was installed recently or is part of standard tooling
', '["https://attack.mitre.org/techniques/T1040/","https://docs.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4688","https://www.sans.org/reading-room/whitepapers/detection/detecting-network-sniffers-1180"]', e'(
  (equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
  (
    regexMatch("log.eventDataNewProcessName", "(?i)(wireshark|tshark|dumpcap|windump|tcpdump|rawcap|netsh\\\\.exe|pktmon\\\\.exe)\\\\.exe$") ||
    regexMatch("log.eventDataNewProcessName", "(?i)(ettercap|NetworkMiner|Microsoft Network Monitor|nmcap)\\\\.exe$") ||
    regexMatch("log.eventDataProcessName", "(?i)(wireshark|tshark|dumpcap|windump|tcpdump|rawcap)\\\\.exe$") ||
    (
      regexMatch("log.eventDataNewProcessName", "(?i)netsh\\\\.exe$") &&
      regexMatch("log.eventDataCommandLine", "(?i)(trace\\\\s+start|capture\\\\s+start)")
    ) ||
    (
      regexMatch("log.eventDataNewProcessName", "(?i)pktmon\\\\.exe$") &&
      contains("log.eventDataCommandLine", "start")
    ) ||
    regexMatch("log.eventDataCommandLine", "(?i)(npcap|winpcap|pcap\\\\.dll|raw\\\\s+socket|promiscuous)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(npcap|winpcap|pcap\\\\.dll|raw\\\\s+socket|promiscuous)")
  )
) ||
(
  (equals("log.eventCode", "4104") || equals("log.eventId", 4104)) &&
  equals("log.providerName", "Microsoft-Windows-PowerShell") &&
  (
    contains("log.eventDataScriptBlockText", "Net.Sockets.Socket") ||
    contains("log.eventDataScriptBlockText", "SocketType.Raw") ||
    contains("log.eventDataScriptBlockText", "ProtocolType.IP") ||
    contains("log.eventDataScriptBlockText", "IOControlCode") ||
    contains("log.eventDataScriptBlockText", "SIO_RCVALL") ||
    contains("log.eventDataScriptBlockText", "Invoke-PacketCapture") ||
    contains("log.eventDataScriptBlockText", "New-NetEventSession")
  )
) ||
(
  equals("log.eventCode", "7045") &&
  regexMatch("log.eventDataServiceName", "(?i)(npcap|winpcap|npf|rawcap)")
)
', '2026-03-02 13:27:39.021892', true, true, 'origin', '["adversary.host","adversary.user"]', '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1070, 'Named Pipe Impersonation Attack', 3, 3, 2, 'Defense Evasion, Privilege Escalation', 'T1134.001 - Access Token Manipulation: Token Impersonation/Theft', e'Detects potential named pipe impersonation attacks used for privilege escalation. Monitors for suspicious process creation patterns including cmd.exe or powershell.exe with pipe-related commands, and processes creating named pipes with suspicious naming patterns commonly used by attack tools like Meterpreter and Cobalt Strike.

Next Steps:
1. Verify the legitimacy of the process and command line parameters
2. Check for any privilege escalation activities following this event
3. Examine the process parent-child relationship and execution context
4. Review recent logon events (4672) for the affected host
5. Investigate any unusual network connections or file access patterns
6. Consider isolating the affected system if malicious activity is confirmed
', '["https://bherunda.medium.com/hunting-named-pipe-token-impersonation-abuse-573dcca36ae0","https://attack.mitre.org/techniques/T1134/001/"]', e'equals("log.eventCode", "4688") &&
exists("log.eventDataProcessName") &&
(contains("log.eventDataProcessName", "cmd.exe") || contains("log.eventDataProcessName", "powershell.exe")) &&
exists("log.eventDataProcessCommandLine") &&
(contains("log.eventDataProcessCommandLine", "\\\\\\\\.\\\\pipe\\\\") ||
 (contains("log.eventDataProcessCommandLine", "echo") && contains("log.eventDataProcessCommandLine", "pipe")) ||
 contains("log.eventDataProcessCommandLine", "CreateNamedPipe") ||
 contains("log.eventDataProcessCommandLine", "ImpersonateNamedPipeClient"))
', '2026-03-02 13:27:40.419829', true, true, 'origin', null, '[]', '["lastEvent.log.eventDataProcessId","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1071, 'MSIExec Remote Payload Installation', 2, 3, 2, 'Defense Evasion', 'T1218.007 - System Binary Proxy Execution: Msiexec', e'Detects abuse of msiexec.exe for installing MSI packages from remote URLs, executing DLLs,
or running packages with quiet/silent flags to avoid user interaction. Attackers use msiexec
as a proxy to execute malicious code because it is a signed Microsoft binary that can download
and execute payloads from remote locations.

Next Steps:
1. Examine the MSI package URL or path for malicious content
2. Review the quiet/silent installation flags for evasion intent
3. Check the parent process to determine the delivery mechanism
4. Analyze the MSI package contents in a sandbox
5. Review network connections to the MSI download source
6. Block the identified URL at the proxy/firewall
7. Search for the same MSI package hash across other endpoints
', '["https://attack.mitre.org/techniques/T1218/007/","https://lolbas-project.github.io/lolbas/Binaries/Msiexec/","https://www.trendmicro.com/en_us/research/19/b/msiexec-abuse.html"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  regexMatch("log.eventDataNewProcessName", "(?i)msiexec\\\\.exe$") ||
  regexMatch("log.eventDataProcessName", "(?i)msiexec\\\\.exe$")
) &&
(
  regexMatch("log.eventDataCommandLine", "(?i)/i\\\\s+(http://|https://)") ||
  regexMatch("log.eventDataCommandLine", "(?i)/y\\\\s+") ||
  regexMatch("log.eventDataCommandLine", "(?i)/z\\\\s+(http://|https://)") ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)(/q|/quiet|/passive)") &&
    regexMatch("log.eventDataCommandLine", "(?i)(http://|https://|\\\\\\\\\\\\\\\\)")
  ) ||
  regexMatch("log.eventDataCommandLine", "(?i)/i\\\\s+(http://|https://)") ||
  regexMatch("log.eventDataCommandLine", "(?i)/y\\\\s+") ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)(/q|/quiet|/passive)") &&
    regexMatch("log.eventDataCommandLine", "(?i)(http://|https://|\\\\\\\\\\\\\\\\)")
  )
)
', '2026-03-02 13:27:41.770652', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1072, 'MSHTA Suspicious Execution Detection', 2, 3, 2, 'Defense Evasion', 'T1218.005 - System Binary Proxy Execution: Mshta', e'Detects suspicious mshta.exe execution patterns including remote HTA file execution, inline
VBScript/JavaScript execution, and COM scriptlet execution. MSHTA is a signed Microsoft binary
that executes Microsoft HTML Applications (HTA) and is frequently abused for initial access,
defense evasion, and payload delivery because it bypasses application whitelisting.

Next Steps:
1. Examine the full command line for remote URLs or inline script content
2. Identify the parent process to determine the initial delivery mechanism
3. Check for downloaded HTA files and analyze their contents
4. Review network connections made by the mshta.exe process
5. Block identified malicious URLs at the proxy/firewall
6. Search for similar mshta executions across other endpoints
7. Check for persistence mechanisms established after mshta execution
', '["https://attack.mitre.org/techniques/T1218/005/","https://lolbas-project.github.io/lolbas/Binaries/Mshta/","https://redcanary.com/threat-detection-report/techniques/mshta/"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  regexMatch("log.eventDataNewProcessName", "(?i)mshta\\\\.exe$") ||
  regexMatch("log.eventDataProcessName", "(?i)mshta\\\\.exe$")
) &&
(
  regexMatch("log.eventDataCommandLine", "(?i)(http://|https://|ftp://)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(vbscript:|javascript:)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(GetObject|Execute|CreateObject|WScript\\\\.Shell)") ||
  regexMatch("log.eventDataCommandLine", "(?i)sct:") ||
  regexMatch("log.eventDataCommandLine", "(?i)\\\\.hta\\\\s") ||
  regexMatch("log.eventDataCommandLine", "(?i)(http://|https://|ftp://)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(vbscript:|javascript:)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(GetObject|Execute|CreateObject|WScript\\\\.Shell)")
)
', '2026-03-02 13:27:43.046761', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1073, 'Mimikatz Tool Usage Detection', 3, 3, 1, 'Credential Access', 'T1003.001 - OS Credential Dumping: LSASS Memory', e'Detects potential Mimikatz credential dumping tool usage through various indicators including characteristic command patterns, LSASS access, and known Mimikatz modules. Mimikatz is a well-known post-exploitation tool used to extract plaintext passwords, hash, PIN code and kerberos tickets from memory.

Next Steps:
1. Immediately isolate the affected host to prevent lateral movement
2. Review the full command line and process execution details
3. Check for any credential theft or privilege escalation activities
4. Examine recent logon events and account usage patterns
5. Scan for additional persistence mechanisms or backdoors
6. Reset passwords for all potentially compromised accounts
7. Review security logs for signs of lateral movement to other systems
8. Conduct forensic analysis of memory dumps and system artifacts
', '["https://attack.mitre.org/techniques/T1003/001/","https://github.com/gentilkiwi/mimikatz","https://docs.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4688"]', e'(
  (equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
  (
    regexMatch("log.eventDataNewProcessName", "(?i)mimikatz") ||
    regexMatch("log.eventDataCommandLine", "(?i)(sekurlsa|kerberos|crypto|vault|lsadump|dpapi)::") ||
    regexMatch("log.eventDataCommandLine", "(?i)(logonpasswords|pth|golden|silver|ticket)") ||
    regexMatch("log.eventDataCommandLine", "(?i)privilege::debug") ||
    contains("log.eventDataCommandLine", "coffee") ||
    contains("log.eventDataCommandLine", "kirbi")
  )
) ||
(
  (equals("log.eventCode", "4104") || equals("log.eventId", 4104)) &&
  equals("log.providerName", "Microsoft-Windows-PowerShell") &&
  (
    regexMatch("log.eventDataScriptBlockText", "(?i)invoke-mimikatz") ||
    regexMatch("log.eventDataScriptBlockText", "(?i)mimikatz\\\\.ps1") ||
    regexMatch("log.eventDataScriptBlockText", "(?i)DumpCreds|DumpCerts") ||
    contains("log.eventDataScriptBlockText", "Win32_ShadowCopy")
  )
) ||
(
  equals("log.eventCode", "10") &&
  regexMatch("log.eventDataTargetImage", "(?i)lsass\\\\.exe") &&
  oneOf("log.eventDataGrantedAccess", ["0x1010", "0x1038", "0x1418", "0x1438"])
)
', '2026-03-02 13:27:44.351658', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1076, 'LSASS Credential Dumping Detection', 3, 3, 1, 'Credential Access', 'T1003.001 - OS Credential Dumping: LSASS Memory', e'Detects multiple LSASS credential dumping techniques beyond Mimikatz, including comsvcs.dll MiniDump,
procdump, nanodump, pypykatz, Task Manager dump, rundll32 with comsvcs.dll, and direct NT API calls.
LSASS holds credentials for all authenticated users and is a primary target for credential theft.
This rule complements the existing Mimikatz detection by covering alternative dumping methods.

Next Steps:
1. Immediately isolate the affected host to prevent lateral movement
2. Identify the dumping technique used and the resulting dump file location
3. Check if the dump file has been exfiltrated to an external destination
4. Review the process tree to understand how the dumping tool was executed
5. Reset passwords for all accounts that were logged into the compromised system
6. Enable Credential Guard or RunAsPPL to protect LSASS
7. Search for the same dumping technique across other endpoints
8. Investigate the initial compromise vector
', '["https://attack.mitre.org/techniques/T1003/001/","https://www.microsoft.com/en-us/security/blog/2022/10/05/detecting-and-preventing-lsass-credential-dumping-attacks/","https://redcanary.com/threat-detection-report/techniques/lsass-memory/"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  (
    regexMatch("log.eventDataCommandLine", "(?i)comsvcs\\\\.dll[,\\\\s]+(MiniDump|#24)") ||
    regexMatch("log.eventDataCommandLine", "(?i)comsvcs\\\\.dll[,\\\\s]+(MiniDump|#24)")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)procdump.*-ma\\\\s+(lsass|\\\\d+)") ||
    regexMatch("log.eventDataCommandLine", "(?i)procdump.*-ma\\\\s+(lsass|\\\\d+)")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)rundll32\\\\.exe.*comsvcs") ||
    regexMatch("log.eventDataCommandLine", "(?i)rundll32\\\\.exe.*comsvcs")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)(nanodump|handlekatz|physmem2profit|dumpert)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(nanodump|handlekatz|physmem2profit|dumpert)")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)taskmgr.*lsass") ||
    regexMatch("log.eventDataCommandLine", "(?i)taskmgr.*lsass")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)(Out-Minidump|Get-Process.*lsass.*\\\\.DumpProcess)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(Out-Minidump|Get-Process.*lsass.*\\\\.DumpProcess)")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)sqldumper\\\\.exe.*lsass") ||
    regexMatch("log.eventDataCommandLine", "(?i)sqldumper\\\\.exe.*lsass")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)createdump\\\\.exe.*(-f|--full)") ||
    regexMatch("log.eventDataCommandLine", "(?i)createdump\\\\.exe.*(-f|--full)")
  )
)
', '2026-03-02 13:27:48.328270', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1077, 'Regsvr32 LOLBIN Abuse Detection', 3, 3, 2, 'Defense Evasion', 'T1218.010 - System Binary Proxy Execution: Regsvr32', e'Detects suspicious regsvr32.exe usage including Squiblydoo attacks (loading remote SCT files),
AppLocker bypass techniques, and execution of DLLs from suspicious locations. Regsvr32 is a
signed Microsoft binary that can be abused to proxy execution of malicious code, bypassing
application whitelisting controls.

Next Steps:
1. Examine the DLL or SCT file being loaded by regsvr32
2. Check if a remote URL was used (Squiblydoo attack indicator)
3. Analyze the parent process that launched regsvr32
4. Review network connections made by the regsvr32 process
5. Check if the loaded DLL is from a suspicious location (Temp, AppData, etc.)
6. Block the remote URL if a network-based attack was used
7. Investigate what payload was delivered through the proxy execution
', '["https://attack.mitre.org/techniques/T1218/010/","https://pentestlab.blog/2017/05/11/applocker-bypass-regsvr32/","https://lolbas-project.github.io/lolbas/Binaries/Regsvr32/"]', e'(equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
(
  regexMatch("log.eventDataNewProcessName", "(?i)regsvr32\\\\.exe$") ||
  regexMatch("log.eventDataProcessName", "(?i)regsvr32\\\\.exe$")
) &&
(
  (regexMatch("log.eventDataCommandLine", "(?i)/s.*/u.*/i:") ||
   regexMatch("log.eventDataCommandLine", "(?i)/s.*/u.*/i:")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)/i:(http|https|ftp)://") ||
   regexMatch("log.eventDataCommandLine", "(?i)/i:(http|https|ftp)://")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)\\\\.sct") ||
   regexMatch("log.eventDataCommandLine", "(?i)\\\\.sct")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\|\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Users\\\\\\\\Public\\\\\\\\)") ||
   regexMatch("log.eventDataCommandLine", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\|\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Users\\\\\\\\Public\\\\\\\\)")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)scrobj\\\\.dll") ||
   regexMatch("log.eventDataCommandLine", "(?i)scrobj\\\\.dll"))
)
', '2026-03-02 13:27:49.634487', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1078, 'LaZagne Credential Harvester Detection', 3, 2, 1, 'Credential Access', 'T1555 - Credentials from Password Stores', e'Detects LaZagne credential harvesting tool execution. LaZagne is an open-source tool that
retrieves credentials stored on a local computer from 25+ sources including browsers, email
clients, databases, Wi-Fi passwords, Windows credentials, and more. The rule monitors for
both the executable name and characteristic command-line arguments used by LaZagne.

Next Steps:
1. Immediately isolate the affected host to prevent credential abuse
2. Identify all credential stores that may have been accessed
3. Review the command-line arguments to determine which modules were used
4. Reset passwords for all accounts stored on the compromised system
5. Check browser credential stores, email clients, and Windows vaults
6. Review Wi-Fi profiles for exposed network credentials
7. Investigate the initial access vector and delivery mechanism
8. Search for LaZagne execution across other endpoints
', '["https://attack.mitre.org/techniques/T1555/","https://github.com/AlessandroZ/LaZagne","https://www.sans.org/blog/detecting-lazagne/"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  regexMatch("log.eventDataNewProcessName", "(?i)lazagne\\\\.exe$") ||
  regexMatch("log.eventDataProcessName", "(?i)lazagne\\\\.exe$") ||
  regexMatch("log.eventDataCommandLine", "(?i)lazagne\\\\s+(all|browsers|chats|databases|games|git|mails|maven|memory|multimedia|php|svn|sysadmin|wifi|windows)") ||
  regexMatch("log.eventDataCommandLine", "(?i)lazagne\\\\s+(all|browsers|chats|databases|games|git|mails|maven|memory|multimedia|php|svn|sysadmin|wifi|windows)") ||
  regexMatch("log.eventDataCommandLine", "(?i)lazagne\\\\.exe") ||
  regexMatch("log.eventDataCommandLine", "(?i)lazagne\\\\.exe")
)
', '2026-03-02 13:27:50.871734', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1079, 'Impacket Lateral Movement Detection', 3, 3, 2, 'Lateral Movement', 'T1021.002 - Remote Services: SMB/Windows Admin Shares', e'Detects Impacket framework lateral movement patterns including wmiexec, smbexec, dcomexec, and atexec.
Impacket is the most commonly used lateral movement framework in real-world attacks. The characteristic
pattern involves cmd.exe /Q /c with output redirection to 127.0.0.1 via named pipes or ADMIN$ share,
as well as specific command-line patterns unique to each Impacket module.

Next Steps:
1. Identify the source of the lateral movement and the compromised credentials used
2. Examine the full command line for data exfiltration or payload delivery
3. Check for Impacket artifacts in the ADMIN$ share or temp directories
4. Review authentication logs for the credential source
5. Isolate affected hosts and block lateral movement paths
6. Reset credentials for all accounts used in the lateral movement
7. Hunt for Impacket usage across all domain-joined systems
8. Review network logs for SMB traffic patterns consistent with Impacket
', '["https://attack.mitre.org/techniques/T1021/002/","https://github.com/fortra/impacket","https://www.13cubed.com/downloads/impacket_exec_commands_cheat_sheet.pdf"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  (
    regexMatch("log.eventDataCommandLine", "(?i)cmd\\\\.exe\\\\s+/Q\\\\s+/c\\\\s+.+\\\\s+1>\\\\s*\\\\\\\\\\\\\\\\127\\\\.0\\\\.0\\\\.1\\\\\\\\") ||
    regexMatch("log.eventDataCommandLine", "(?i)cmd\\\\.exe\\\\s+/Q\\\\s+/c\\\\s+.+\\\\s+1>\\\\s*\\\\\\\\\\\\\\\\127\\\\.0\\\\.0\\\\.1\\\\\\\\")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)cmd\\\\.exe\\\\s+/Q\\\\s+/c\\\\s+.+\\\\s+2>&1") ||
    regexMatch("log.eventDataCommandLine", "(?i)cmd\\\\.exe\\\\s+/Q\\\\s+/c\\\\s+.+\\\\s+2>&1")
  ) ||
  (
    regexMatch("log.eventDataParentProcessName", "(?i)wmiprvse\\\\.exe$") &&
    regexMatch("log.eventDataCommandLine", "(?i)cmd\\\\.exe\\\\s+/Q\\\\s+/c")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)\\\\\\\\__output\\\\s") ||
    regexMatch("log.eventDataCommandLine", "(?i)\\\\\\\\__output\\\\s")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)cmd\\\\.exe\\\\s+/C\\\\s+.+\\\\\\\\ADMIN\\\\$\\\\\\\\__\\\\d+\\\\.\\\\d+") ||
    regexMatch("log.eventDataCommandLine", "(?i)cmd\\\\.exe\\\\s+/C\\\\s+.+\\\\\\\\ADMIN\\\\$\\\\\\\\__\\\\d+\\\\.\\\\d+")
  )
)
', '2026-03-02 13:27:52.044424', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1080, 'Image File Execution Options Debugger Persistence', 3, 3, 2, 'Persistence', 'T1546.012 - Event Triggered Execution: Image File Execution Options Injection', e'Detects abuse of Image File Execution Options (IFEO) registry keys to establish persistence
or redirect program execution. Attackers set the Debugger value under IFEO to hijack
execution of legitimate programs (like accessibility tools sethc.exe, utilman.exe, narrator.exe)
and redirect them to cmd.exe or other malicious payloads. This technique is also used for
persistence through GlobalFlag and SilentProcessExit monitoring.

Next Steps:
1. Check which program\'s IFEO key was modified and what debugger was set
2. Verify if accessibility tool hijacking was used (sethc, utilman, narrator, magnify, osk)
3. Remove the malicious Debugger registry value
4. Check for RDP access if accessibility tools were hijacked (sticky keys attack)
5. Investigate the user account that made the registry modification
6. Search for additional persistence mechanisms
7. Review recent RDP login activity on the affected host
', '["https://attack.mitre.org/techniques/T1546/012/","https://blog.malwarebytes.com/101/2015/12/an-introduction-to-image-file-execution-options/","https://oddvar.moe/2018/04/10/persistence-using-globalflags-in-image-file-execution-options/"]', e'(
  equals("log.eventCode", "13") &&
  equals("log.providerName", "Microsoft-Windows-Sysmon") &&
  regexMatch("log.eventDataTargetObject", "(?i)\\\\\\\\Image File Execution Options\\\\\\\\.*\\\\\\\\(Debugger|GlobalFlag|MonitorProcess)$") &&
  !regexMatch("log.eventDataDetails", "(?i)(werfault|drwtsn32|vsjitdebugger|windbg)")
) ||
(
  (equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
  (
    regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+add.*Image File Execution Options.*Debugger") ||
    regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+add.*Image File Execution Options.*Debugger")
  )
)
', '2026-03-02 13:27:53.136703', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1081, 'GPO Tampering Detection', 2, 3, 2, 'Defense Evasion, Privilege Escalation', 'T1484.001 - Domain Policy Modification: Group Policy Modification', e'Detects modifications to Group Policy Objects (GPOs) which could indicate an adversary attempting to escalate privileges, deploy malware across the domain, or establish persistence. GPO tampering is a common technique used by attackers to enforce malicious configurations or deploy payloads to multiple systems simultaneously.

**Next Steps:**
1. Immediately review the specific GPO that was modified and identify what changes were made
2. Verify if the modification was authorized by checking with the responsible administrator
3. Examine the user account that made the changes for signs of compromise
4. Review recent authentication logs for the user account to identify potential lateral movement
5. Check domain controllers for additional suspicious activities around the same timeframe
6. If unauthorized, immediately revert the GPO changes and investigate the compromise vector
7. Consider temporarily disabling the affected user account pending investigation
', '["https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/component-updates/command-line-process-auditing","https://attack.mitre.org/techniques/T1484/001/"]', 'equals("log.eventId", "5136") && equals("log.providerName", "Microsoft-Windows-Security-Auditing") && contains("log.eventDataObjectDN", "CN=Policies,CN=System")', '2026-03-02 13:27:54.184266', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1082, 'Event Log Clearing Detection', 3, 3, 2, 'Defense Evasion', 'T1070.001 - Indicator Removal on Host: Clear Windows Event Logs', e'Detects when Windows event logs are cleared, which is often done by attackers to cover their tracks and remove evidence of malicious activities. This rule monitors for Event ID 1102 (Security log cleared), Event ID 104 (System log cleared), and command-line activities using wevtutil or PowerShell cmdlets to clear event logs.

Next Steps:
1. Identify the user account that performed the log clearing operation
2. Review the timeline of events before the log clearing to identify potential malicious activities
3. Check for any remaining forensic artifacts in other log sources (network logs, endpoint logs, etc.)
4. Investigate if this was authorized maintenance or potential malicious activity
5. Review user privileges and access patterns for the account involved
6. Consider implementing additional logging and monitoring for critical systems
', '["https://attack.mitre.org/techniques/T1070/001/","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventid=1102"]', e'equals("log.eventId", "1102") ||
equals("log.eventCode", "1102") ||
(equals("log.eventId", "104") &&
 contains("log.channel", "System")) ||
(equals("log.eventId", "4688") &&
 ((contains("log.eventDataCommandLine", "wevtutil") &&
   contains("log.eventDataCommandLine", " cl ")) ||
  contains("log.eventDataCommandLine", "Clear-EventLog") ||
  contains("log.eventDataCommandLine", "Remove-EventLog") ||
  contains("log.eventDataCommandLine", "Limit-EventLog")))
', '2026-03-02 13:27:55.373043', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1083, 'ETW Patching and Tampering Detection', 3, 3, 2, 'Defense Evasion', 'T1562.006 - Impair Defenses: Indicator Blocking', e'Detects attempts to tamper with Event Tracing for Windows (ETW) to blind security tools.
Attackers patch ETW functions in memory (NtTraceEvent, EtwEventWrite) or use PowerShell
Reflection to disable .NET ETW providers, preventing security tools from receiving telemetry.
This technique is frequently used before executing post-exploitation tools to evade detection.

Next Steps:
1. Investigate the process that attempted ETW tampering
2. Check what malicious activity followed the ETW patching
3. Review PowerShell script block logs for ETW manipulation code
4. Examine if the .NET ETW provider was targeted (common for AMSI bypass chains)
5. Verify that ETW providers are functioning correctly after remediation
6. Search for additional evasion techniques on the same host
7. Isolate the host and investigate the full attack chain
', '["https://attack.mitre.org/techniques/T1562/006/","https://blog.xpnsec.com/hiding-your-dotnet-etw/","https://www.mdsec.co.uk/2020/03/hiding-your-net-etw/"]', e'(
  (equals("log.eventCode", "4104") || equals("log.eventCode", "4103")) &&
  equals("log.providerName", "Microsoft-Windows-PowerShell") &&
  (
    contains("log.eventDataScriptBlockText", "EtwEventWrite") ||
    contains("log.eventDataScriptBlockText", "NtTraceEvent") ||
    contains("log.eventDataScriptBlockText", "NtTraceControl") ||
    regexMatch("log.eventDataScriptBlockText", "(?i)\\\\[Reflection\\\\.Assembly\\\\].*ETW") ||
    regexMatch("log.eventDataScriptBlockText", "(?i)Patch.*ETW|ETW.*Patch") ||
    contains("log.eventDataScriptBlockText", "EventProvider") ||
    contains("log.eventDataScriptBlockText", "EtwEventWrite") ||
    contains("log.eventDataScriptBlockText", "NtTraceEvent")
  )
) ||
(
  (equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
  (
    regexMatch("log.eventDataCommandLine", "(?i)logman.*stop.*EventLog") ||
    regexMatch("log.eventDataCommandLine", "(?i)logman.*stop.*EventLog") ||
    regexMatch("log.eventDataCommandLine", "(?i)auditpol.*/clear") ||
    regexMatch("log.eventDataCommandLine", "(?i)auditpol.*/clear") ||
    regexMatch("log.eventDataCommandLine", "(?i)Set-EtwTraceProvider.*0x0") ||
    regexMatch("log.eventDataCommandLine", "(?i)Set-EtwTraceProvider.*0x0")
  )
)
', '2026-03-02 13:27:56.633563', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1084, 'DPAPI Domain Backup Key Extraction', 3, 3, 1, 'Credential Access', 'T1003.004 - OS Credential Dumping: LSA Secrets', e'Detects extraction of the DPAPI domain backup key from a domain controller. The BCKUPKEY secret
object contains the domain\'s DPAPI backup key which can be used to decrypt any DPAPI-protected
data for all domain users, including passwords, certificates, and other sensitive data. This is
a high-severity credential access technique that indicates an attacker has domain admin access.

Next Steps:
1. Verify the account accessing the BCKUPKEY object is authorized (should only be DC replication)
2. Check if the source host is a legitimate domain controller
3. Review the user account for signs of compromise
4. Investigate for related DCSync or NTDS.dit extraction attempts
5. Assume all DPAPI-protected secrets are compromised if unauthorized
6. Rotate the DPAPI domain backup key (requires careful planning)
7. Reset credentials for all privileged accounts
', '["https://attack.mitre.org/techniques/T1003/004/","https://www.dsinternals.com/en/retrieving-dpapi-backup-keys-from-active-directory/","https://adsecurity.org/?p=1785"]', e'equals("log.eventCode", "4662") &&
equals("log.channel", "Security") &&
contains("log.eventDataProperties", "BCKUPKEY") &&
contains("log.eventDataObjectServer", "DS") &&
!regexMatch("log.eventDataSubjectUserName", "(?i)\\\\$$")
', '2026-03-02 13:27:57.757884', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1085, 'Domain Trust Discovery Detection', 2, 1, 1, 'Discovery', 'T1482 - Domain Trust Discovery', e'Detects domain trust enumeration using nltest, dsquery, Get-ADTrust, and other Active
Directory discovery tools. Attackers enumerate domain trusts to identify additional targets
for lateral movement between trusted domains and forests. This is typically an early
reconnaissance step in Active Directory attacks.

Next Steps:
1. Identify the user account performing domain trust enumeration
2. Verify if this is authorized security assessment or IT administration
3. Check for other discovery commands from the same user or host
4. Review if the user subsequently accessed resources in trusted domains
5. Correlate with other reconnaissance activities (BloodHound, SharpHound)
6. Monitor for lateral movement into discovered trusted domains
7. Review trust configurations for unnecessary or overly permissive trusts
', '["https://attack.mitre.org/techniques/T1482/","https://adsecurity.org/?p=1588","https://www.harmj0y.net/blog/redteaming/a-guide-to-attacking-domain-trusts/"]', e'(equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
(
  (regexMatch("log.eventDataCommandLine", "(?i)nltest(.exe)?\\\\s+/(domain_trusts|trusted_domains|dclist|dsgetdc)") ||
   regexMatch("log.eventDataCommandLine", "(?i)nltest(.exe)?\\\\s+/(domain_trusts|trusted_domains|dclist|dsgetdc)")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)dsquery(.exe)?\\\\s+trust") ||
   regexMatch("log.eventDataCommandLine", "(?i)dsquery(.exe)?\\\\s+trust")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)Get-ADTrust|Get-DomainTrust|Get-ForestTrust|Get-NetForestTrust") ||
   regexMatch("log.eventDataCommandLine", "(?i)Get-ADTrust|Get-DomainTrust|Get-ForestTrust|Get-NetForestTrust")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)\\\\[System\\\\.DirectoryServices\\\\.ActiveDirectory\\\\.(Domain|Forest)\\\\]") ||
   regexMatch("log.eventDataCommandLine", "(?i)\\\\[System\\\\.DirectoryServices\\\\.ActiveDirectory\\\\.(Domain|Forest)\\\\]")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)adfind(.exe)?.*-f.*trustdirection") ||
   regexMatch("log.eventDataCommandLine", "(?i)adfind(.exe)?.*-f.*trustdirection"))
)
', '2026-03-02 13:27:58.951717', true, true, 'origin', '["adversary.host","adversary.user"]', '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1086, 'DLL Sideloading Detection', 2, 3, 1, 'Defense Evasion', 'T1574.002 - Hijack Execution Flow: DLL Side-Loading', e'Detects DLL sideloading patterns where known vulnerable legitimate applications load malicious
DLLs from non-system paths. Attackers place a malicious DLL alongside a vulnerable application
that loads DLLs from its own directory rather than the system directory. This technique
leverages trusted application signatures to execute malicious code and evade security controls.

Next Steps:
1. Verify the executable path - legitimate apps should be in Program Files, not user directories
2. Check the DLL loaded alongside the executable for unexpected modifications
3. Compare DLL hashes against known good versions
4. Examine the parent process to understand how the vulnerable app was launched
5. Review file creation timestamps for the executable and DLL pair
6. Analyze the suspicious DLL in a sandbox
7. Search for similar sideloading patterns across other endpoints
', '["https://attack.mitre.org/techniques/T1574/002/","https://www.mandiant.com/resources/blog/dll-side-loading-a-thorn-in-the-side-of-the-anti-virus-industry","https://hijacklibs.net/"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  (
    regexMatch("log.eventDataNewProcessName", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\|\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Desktop\\\\\\\\|\\\\\\\\Users\\\\\\\\Public\\\\\\\\)") &&
    regexMatch("log.eventDataNewProcessName", "(?i)(OneDriveStandaloneUpdater|colorcpl|consent|dxcap|eudcedit|eventvwr|isoburn|msconfig|msdt|mstsc|narrator|netplwiz|odbcad32|presentationhost|rstrui|sdclt|sethc|sigverif|utilman|write)\\\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataProcessName", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\|\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Desktop\\\\\\\\)") &&
    regexMatch("log.eventDataProcessName", "(?i)(OneDriveStandaloneUpdater|colorcpl|consent|dxcap|eudcedit|eventvwr|isoburn|msconfig|msdt|mstsc|narrator|netplwiz|odbcad32|presentationhost|rstrui|sdclt|sethc|sigverif|utilman|write)\\\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataNewProcessName", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\|\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Desktop\\\\\\\\|\\\\\\\\Users\\\\\\\\Public\\\\\\\\)") &&
    regexMatch("log.eventDataNewProcessName", "(?i)(WerFault|SearchProtocolHost|SearchFilterHost|WmiPrvSE|backgroundTaskHost|RuntimeBroker|smartscreen|tabcal|winsat)\\\\.exe$")
  )
)
', '2026-03-02 13:28:00.829760', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1087, 'DCOM Lateral Movement Detection', 3, 3, 2, 'Lateral Movement', 'T1021.003 - Remote Services: Distributed Component Object Model', e'Detects potential DCOM lateral movement attempts by monitoring for suspicious process creation with DCOM-related command line parameters. Looks for processes with automation embedding flags and specific DCOM object CLSIDs commonly abused for lateral movement such as ShellWindows and MMC20.Application.

Next Steps:
1. Investigate the source host and user account that initiated the DCOM activity
2. Review network connections between the source and target systems around the time of the alert
3. Check for additional lateral movement indicators on both source and target systems
4. Examine the specific DCOM objects and CLSIDs being accessed for known malicious usage
5. Verify if the DCOM activity aligns with legitimate business processes or scheduled tasks
6. Look for privilege escalation attempts following the lateral movement
7. Check for any data exfiltration or persistence mechanisms deployed post-compromise
', '["https://medium.com/@cY83rR0H1t/detecting-dcom-lateral-movement-ee2b461a8705","https://attack.mitre.org/techniques/T1021/003/"]', 'equals("log.eventCode", "4688") && exists("log.eventDataProcessCommandLine") && (contains("log.eventDataProcessCommandLine", "/automation -Embedding") || contains("log.eventDataProcessCommandLine", "9BA05972-F6A8-11CF-A442-00A0C90A8F39") || contains("log.eventDataProcessCommandLine", "c08afd90-f2a1-11d1-8455-00a0c91f3880") || contains("log.eventDataProcessCommandLine", "MMC20.Application") || contains("log.eventDataProcessCommandLine", "Document.Application.ShellExecute") || contains("log.eventDataProcessCommandLine", "GetTypeFromCLSID") || contains("log.eventDataProcessCommandLine", "GetTypeFromProgID"))', '2026-03-02 13:28:02.045113', true, true, 'origin', null, '[]', '["lastEvent.log.eventDataProcessName","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1088, 'Credential Manager Access Detection', 3, 2, 1, 'Credential Access', 'T1555.004 - Credentials from Password Stores: Windows Credential Manager', e'Detects access to Windows Credential Manager using vaultcmd.exe, cmdkey.exe, or credential
enumeration via PowerShell. Attackers use these tools to extract stored credentials including
web passwords, Windows credentials, and RDP saved logins. This is a common post-exploitation
technique for credential harvesting.

Next Steps:
1. Identify the user account and process executing credential manager commands
2. Verify if this is authorized administrative activity or penetration testing
3. Review what credentials are stored in the affected user\'s vault
4. Check for follow-up lateral movement using extracted credentials
5. Reset any credentials that may have been exposed
6. Review if remote desktop saved credentials were compromised
7. Investigate the initial access vector that led to credential access
', '["https://attack.mitre.org/techniques/T1555/004/","https://www.passcape.com/index.php?section=docsys&cmd=details&id=28","https://blog.malwarebytes.com/threat-analysis/2020/12/credential-stealing/"]', e'(equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
(
  (regexMatch("log.eventDataCommandLine", "(?i)vaultcmd(.exe)?\\\\s+/(list|listcreds|listproperties)") ||
   regexMatch("log.eventDataCommandLine", "(?i)vaultcmd(.exe)?\\\\s+/(list|listcreds|listproperties)")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)cmdkey(.exe)?\\\\s+/list") ||
   regexMatch("log.eventDataCommandLine", "(?i)cmdkey(.exe)?\\\\s+/list")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)(CredentialManager|Windows\\\\s+Vault|VaultSvc)") ||
   regexMatch("log.eventDataCommandLine", "(?i)(CredentialManager|Windows\\\\s+Vault|VaultSvc)")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)Get-VaultCredential|Get-CachedGPPPassword") ||
   regexMatch("log.eventDataCommandLine", "(?i)Get-VaultCredential|Get-CachedGPPPassword"))
)
', '2026-03-02 13:28:03.316781', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1089, 'Credential Dump via Registry Export', 3, 2, 1, 'Credential Access', 'T1003.004 - OS Credential Dumping: LSA Secrets', e'Detects credential dumping via registry export of SAM, SYSTEM, and SECURITY hives using
reg.exe save commands. Attackers export these hives to extract password hashes offline using
tools like secretsdump.py or samdump2. This technique is commonly used after gaining local
administrator privileges.

Next Steps:
1. Identify the user account executing the reg save commands
2. Check if SAM, SYSTEM, and SECURITY hives were all exported (indicates deliberate extraction)
3. Search for the output files on disk and check if they were exfiltrated
4. Review for follow-up pass-the-hash or credential reuse activity
5. Reset all local account passwords on the affected system
6. Investigate how the attacker obtained local administrator privileges
7. Check for domain credential exposure if SECURITY hive was exported
', '["https://attack.mitre.org/techniques/T1003/004/","https://www.sans.org/blog/protecting-privileged-domain-accounts-safeguarding-password-hashes/","https://pentestlab.blog/2018/07/04/dumping-domain-password-hashes/"]', e'(equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
(
  (regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+(save|export)\\\\s+[\\"\']?HKLM\\\\\\\\SAM") ||
   regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+(save|export)\\\\s+[\\"\']?HKLM\\\\\\\\SAM")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+(save|export)\\\\s+[\\"\']?HKLM\\\\\\\\SYSTEM") ||
   regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+(save|export)\\\\s+[\\"\']?HKLM\\\\\\\\SYSTEM")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+(save|export)\\\\s+[\\"\']?HKLM\\\\\\\\SECURITY") ||
   regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+(save|export)\\\\s+[\\"\']?HKLM\\\\\\\\SECURITY")) ||
  (regexMatch("log.eventDataCommandLine", "(?i)esentutl.*/y.*/d.*ntds\\\\.dit") ||
   regexMatch("log.eventDataCommandLine", "(?i)esentutl.*/y.*/d.*ntds\\\\.dit"))
)
', '2026-03-02 13:28:04.757706', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1090, 'COM Hijacking Persistence Detection', 2, 3, 1, 'Persistence', 'T1546.015 - Event Triggered Execution: Component Object Model Hijacking', e'Detects COM (Component Object Model) hijacking used for persistence by modifying InProcServer32
or LocalServer32 registry values. Attackers replace legitimate COM object DLLs with malicious ones
to achieve persistence, as the malicious DLL will be loaded whenever the COM object is instantiated.
This is a stealthy persistence mechanism that is difficult to detect without registry monitoring.

Next Steps:
1. Examine the modified InProcServer32/LocalServer32 value to identify the malicious DLL
2. Verify if the new DLL path points to a legitimate or suspicious location
3. Check the original COM object registration for comparison
4. Analyze the replacement DLL in a sandbox
5. Identify the CLSID being hijacked and what software uses it
6. Restore the original COM registration
7. Search for similar COM hijacking across other endpoints
', '["https://attack.mitre.org/techniques/T1546/015/","https://pentestlab.blog/2020/05/20/persistence-com-hijacking/","https://bohops.com/2018/08/18/abusing-the-com-registry-structure-clsid-localserver32-inprocserver32/"]', e'equals("log.eventCode", "4657") &&
equals("log.channel", "Security") &&
regexMatch("log.eventDataObjectName", "(?i)(InProcServer32|LocalServer32)") &&
(
  regexMatch("log.eventDataNewValue", "(?i)(\\\\\\\\Temp\\\\\\\\|\\\\\\\\Downloads\\\\\\\\|\\\\\\\\AppData\\\\\\\\|\\\\\\\\ProgramData\\\\\\\\|\\\\\\\\Users\\\\\\\\Public\\\\\\\\|\\\\\\\\Desktop\\\\\\\\)") ||
  regexMatch("log.eventDataNewValue", "(?i)(rundll32|regsvr32|mshta|powershell|cmd\\\\.exe|wscript|cscript)") ||
  regexMatch("log.eventDataNewValue", "(?i)scrobj\\\\.dll") ||
  regexMatch("log.eventDataNewValue", "(?i)(http://|https://)") ||
  !regexMatch("log.eventDataNewValue", "(?i)^(C:\\\\\\\\Windows\\\\\\\\|C:\\\\\\\\Program Files)")
) &&
!regexMatch("log.eventDataSubjectUserName", "(?i)^(SYSTEM|TrustedInstaller)$")
', '2026-03-02 13:28:06.114345', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1091, 'Cobalt Strike Process Behavior Detection', 3, 3, 2, 'Execution', 'T1059 - Command and Scripting Interpreter', e'Detects Cobalt Strike beacon process creation patterns that are distinct from normal system behavior.
These include rundll32.exe spawned without command-line arguments, dllhost.exe or runonce.exe spawning
cmd.exe or powershell, and characteristic post-exploitation process injection patterns. Cobalt Strike
is the most widely used commercial adversary simulation tool and is frequently found in real attacks.

Next Steps:
1. Examine the parent-child process relationships for injection patterns
2. Check for rundll32 with no arguments (classic beacon default)
3. Review named pipe activity on the host for CS pipe names
4. Check for network beaconing behavior from the suspicious process
5. Memory scan the suspicious processes for CS beacon shellcode
6. Isolate the affected host immediately
7. Hunt for lateral movement from this host to other systems
8. Review the C2 infrastructure and block at network perimeter
', '["https://attack.mitre.org/software/S0154/","https://thedfirreport.com/2021/08/29/cobalt-strike-a-defenders-guide/","https://redcanary.com/threat-detection-report/threats/cobalt-strike/"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  (
    regexMatch("log.eventDataNewProcessName", "(?i)rundll32\\\\.exe$") &&
    (
      !exists("log.eventDataCommandLine") ||
      regexMatch("log.eventDataCommandLine", "(?i)^\\"?[A-Z]:\\\\\\\\Windows\\\\\\\\(System32|SysWOW64)\\\\\\\\rundll32\\\\.exe\\"?\\\\s*$")
    )
  ) ||
  (
    regexMatch("log.eventDataParentProcessName", "(?i)(dllhost\\\\.exe|runonce\\\\.exe|searchprotocolhost\\\\.exe)$") &&
    regexMatch("log.eventDataNewProcessName", "(?i)(cmd\\\\.exe|powershell\\\\.exe|pwsh\\\\.exe)$")
  ) ||
  (
    regexMatch("log.eventDataParentProcessName", "(?i)rundll32\\\\.exe$") &&
    regexMatch("log.eventDataNewProcessName", "(?i)(cmd\\\\.exe|powershell\\\\.exe|pwsh\\\\.exe)$") &&
    !exists("log.eventDataParentCommandLine")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)(MSSE-\\\\d+-server|status_\\\\d+|postex_\\\\d+|msagent_\\\\d+)") ||
    regexMatch("log.eventDataCommandLine", "(?i)(MSSE-\\\\d+-server|status_\\\\d+|postex_\\\\d+|msagent_\\\\d+)")
  ) ||
  (
    regexMatch("log.eventDataCommandLine", "(?i)\\\\\\\\\\\\\\\\\\\\.\\\\\\\\.pipe\\\\\\\\(MSSE-|postex_|status_|msagent_)") ||
    regexMatch("log.eventDataCommandLine", "(?i)\\\\\\\\\\\\\\\\\\\\.\\\\\\\\.pipe\\\\\\\\(MSSE-|postex_|status_|msagent_)")
  )
)
', '2026-03-02 13:28:07.383219', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1092, 'CMSTP UAC Bypass Detection', 2, 3, 1, 'Defense Evasion', 'T1218.003 - System Binary Proxy Execution: CMSTP', e'Detects CMSTP.exe (Microsoft Connection Manager Profile Installer) being used to bypass UAC
and AppLocker restrictions. Attackers use CMSTP with specially crafted .inf files containing
malicious commands in the RunPreSetupCommandsSection to execute arbitrary code with elevated
privileges. This is a well-known UAC bypass technique.

Next Steps:
1. Examine the .inf file referenced in the command line for malicious content
2. Check the RunPreSetupCommandsSection of the INF file for commands
3. Identify the parent process and delivery mechanism
4. Review if UAC was successfully bypassed
5. Check for post-exploitation activity with elevated privileges
6. Remove the malicious INF file and any created artifacts
7. Search for similar CMSTP abuse across other endpoints
', '["https://attack.mitre.org/techniques/T1218/003/","https://lolbas-project.github.io/lolbas/Binaries/Cmstp/","https://msitpros.com/?p=3960"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  regexMatch("log.eventDataNewProcessName", "(?i)cmstp\\\\.exe$") ||
  regexMatch("log.eventDataProcessName", "(?i)cmstp\\\\.exe$")
) &&
(
  regexMatch("log.eventDataCommandLine", "(?i)/s\\\\s+.*\\\\.inf") ||
  regexMatch("log.eventDataCommandLine", "(?i)/ni\\\\s+/s\\\\s+") ||
  regexMatch("log.eventDataCommandLine", "(?i)/au\\\\s+") ||
  regexMatch("log.eventDataCommandLine", "(?i)/s\\\\s+.*\\\\.inf") ||
  regexMatch("log.eventDataCommandLine", "(?i)/ni\\\\s+/s\\\\s+") ||
  regexMatch("log.eventDataCommandLine", "(?i)/au\\\\s+")
)
', '2026-03-02 13:28:08.599372', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1093, 'Certutil LOLBIN Abuse Detection', 2, 3, 1, 'Defense Evasion', 'T1105 - Ingress Tool Transfer', e'Detects abuse of certutil.exe as a Living Off The Land Binary (LOLBIN) for downloading files from URLs,
encoding/decoding Base64 payloads, and NTLM coercion. Certutil is a legitimate Windows certificate
utility that is frequently abused by attackers for payload staging and defense evasion because it is
a signed Microsoft binary that bypasses application whitelisting.

Next Steps:
1. Examine the full command line to identify downloaded URLs or encoded payloads
2. Check the destination file path for downloaded or decoded files
3. Analyze any downloaded files in a sandbox environment
4. Review the parent process to understand how certutil was invoked
5. Check for subsequent execution of downloaded payloads
6. Block the identified download URLs at the proxy/firewall level
7. Search for similar certutil abuse across other endpoints
', '["https://attack.mitre.org/techniques/T1105/","https://attack.mitre.org/techniques/T1140/","https://lolbas-project.github.io/lolbas/Binaries/Certutil/"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  regexMatch("log.eventDataNewProcessName", "(?i)certutil\\\\.exe$") ||
  regexMatch("log.eventDataProcessName", "(?i)certutil\\\\.exe$")
) &&
(
  regexMatch("log.eventDataCommandLine", "(?i)(-urlcache|-URL)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(-encode|-decode)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(-ping|-generateSSTFromWU)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(http://|https://|ftp://)") ||
  regexMatch("log.eventDataCommandLine", "(?i)-verifyctl") ||
  regexMatch("log.eventDataCommandLine", "(?i)(-urlcache|-URL)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(-encode|-decode)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(http://|https://|ftp://)")
)
', '2026-03-02 13:28:09.949837', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1094, 'Certificate Services Abuse Detection', 3, 3, 1, 'Credential Access', 'T1558 - Steal or Forge Kerberos Tickets', e'Detects suspicious certificate requests and issuance that could indicate Golden Certificate attacks or unauthorized certificate generation for persistence. This rule monitors Windows Certificate Services events for potentially malicious certificate operations, particularly those involving machine accounts or anonymous logons that could be leveraged for persistence and privilege escalation.

Next Steps:
1. Investigate the certificate request details including the requesting user/machine
2. Verify if the certificate request was legitimate and authorized
3. Check for any recent changes to Certificate Authority policies or templates
4. Review Certificate Authority logs for other suspicious certificate issuance
5. Examine the requesting host for signs of compromise
6. Consider revoking any suspicious certificates issued
7. Validate Certificate Authority security configurations and access controls
', '["https://www.splunk.com/en_us/blog/security/breaking-the-chain-defending-against-certificate-services-abuse.html","https://attack.mitre.org/techniques/T1558/"]', '(equals("log.eventId", "4886") || equals("log.eventId", "4887")) && equals("log.providerName", "Microsoft-Windows-Security-Auditing") && (contains("log.eventDataSubjectUserName", "$") || equals("log.eventDataSubjectUserName", "ANONYMOUS LOGON"))', '2026-03-02 13:28:11.218232', true, true, 'origin', null, '[]', '["lastEvent.log.eventDataSubjectUserName","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1095, 'Boot and Logon Script Persistence Detection', 2, 3, 2, 'Persistence', 'T1037.001 - Boot or Logon Initialization Scripts: Logon Script (Windows)', e'Detects modifications to Windows logon script registry keys including UserInitMprLogonScript,
Userinit, and Shell values. Attackers modify these keys to execute malicious scripts or
binaries during user logon, providing persistent access that survives reboots. These keys
are commonly abused because they execute with the logged-on user\'s privileges.

Next Steps:
1. Examine the registry value data to identify the script or binary being persisted
2. Analyze the referenced script or executable for malicious content
3. Verify the parent process that modified the registry key
4. Restore the original Userinit or Shell values
5. Remove any malicious scripts from the referenced paths
6. Search for additional persistence mechanisms on the same host
7. Investigate the initial compromise vector
', '["https://attack.mitre.org/techniques/T1037/001/","https://www.cybereason.com/blog/persistence-techniques-that-persist","https://pentestlab.blog/2020/01/14/persistence-logon-scripts/"]', e'(
  equals("log.eventCode", "13") &&
  equals("log.providerName", "Microsoft-Windows-Sysmon") &&
  (
    regexMatch("log.eventDataTargetObject", "(?i)\\\\\\\\Windows\\\\\\\\CurrentVersion\\\\\\\\(on\\\\\\\\(Userinit|Shell)|Policies\\\\\\\\Explorer\\\\\\\\Run)") ||
    regexMatch("log.eventDataTargetObject", "(?i)UserInitMprLogonScript$") ||
    regexMatch("log.eventDataTargetObject", "(?i)\\\\\\\\Environment\\\\\\\\UserInitMprLogonScript$")
  ) &&
  !regexMatch("log.eventDataDetails", "(?i)^C:\\\\\\\\Windows\\\\\\\\system32\\\\\\\\userinit\\\\.exe,?\\\\s*$") &&
  !regexMatch("log.eventDataDetails", "(?i)^explorer\\\\.exe\\\\s*$")
) ||
(
  (equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
  (
    regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+add.*on.*(Userinit|Shell)") ||
    regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+add.*on.*(Userinit|Shell)") ||
    regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+add.*UserInitMprLogonScript") ||
    regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+add.*UserInitMprLogonScript")
  )
)
', '2026-03-02 13:28:12.540759', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1096, 'BITSAdmin Abuse Detection', 2, 3, 1, 'Defense Evasion', 'T1197 - BITS Jobs', e'Detects abuse of bitsadmin.exe for downloading files from remote URLs, creating persistent BITS
jobs, and transferring payloads. BITS (Background Intelligent Transfer Service) is a Windows
service commonly abused by attackers for both download and persistence because BITS jobs survive
reboots and can be configured to execute commands upon completion.

Next Steps:
1. Examine the download URL and destination path in the command line
2. Review the BITS job for any notification command (persistence mechanism)
3. Check the downloaded file for malicious content
4. List all BITS jobs on the system using \'bitsadmin /list /allusers /verbose\'
5. Remove suspicious BITS jobs and quarantine downloaded files
6. Block the download URL at the proxy/firewall level
7. Search for similar BITS abuse across other endpoints
', '["https://attack.mitre.org/techniques/T1197/","https://lolbas-project.github.io/lolbas/Binaries/Bitsadmin/","https://isc.sans.edu/diary/Investigating+Microsoft+BITS+Activity/23281"]', e'(equals("log.eventCode", "4688") || equals("log.eventId", 4688)) &&
(
  regexMatch("log.eventDataNewProcessName", "(?i)bitsadmin\\\\.exe$") ||
  regexMatch("log.eventDataProcessName", "(?i)bitsadmin\\\\.exe$")
) &&
(
  regexMatch("log.eventDataCommandLine", "(?i)(/transfer|/addfile|/resume|/create|/setnotifycmdline|/setnotifyflags)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(http://|https://|ftp://)") ||
  regexMatch("log.eventDataCommandLine", "(?i)/SetMinRetryDelay") ||
  regexMatch("log.eventDataCommandLine", "(?i)(/transfer|/addfile|/resume|/create|/setnotifycmdline)") ||
  regexMatch("log.eventDataCommandLine", "(?i)(http://|https://|ftp://)")
)
', '2026-03-02 13:28:13.935398', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1097, 'AppInit DLLs Persistence Detection', 3, 3, 2, 'Persistence', 'T1546.010 - Event Triggered Execution: AppInit DLLs', e'Detects modifications to the AppInit_DLLs registry key, which causes specified DLLs to be
loaded into every process that loads User32.dll. Attackers use this technique for persistence
and DLL injection, as the malicious DLL will be loaded into virtually every user-mode process.
On modern Windows with Secure Boot, this requires LoadAppInit_DLLs to also be enabled.

Next Steps:
1. Identify the DLL path being added to the AppInit_DLLs registry value
2. Analyze the referenced DLL for malicious functionality
3. Check if LoadAppInit_DLLs was also enabled (required on modern Windows)
4. Remove the malicious DLL path from the registry value
5. Delete the malicious DLL file from disk
6. Reboot the system to stop the DLL from being loaded into new processes
7. Investigate the initial compromise that led to this persistence
', '["https://attack.mitre.org/techniques/T1546/010/","https://docs.microsoft.com/en-us/windows/win32/dlls/secure-boot-and-appinit-dlls","https://pentestlab.blog/2020/01/07/persistence-appinit-dlls/"]', e'(
  equals("log.eventCode", "13") &&
  equals("log.providerName", "Microsoft-Windows-Sysmon") &&
  regexMatch("log.eventDataTargetObject", "(?i)\\\\\\\\Windows\\\\\\\\CurrentVersion\\\\\\\\Windows\\\\\\\\(AppInit_DLLs|LoadAppInit_DLLs)$") &&
  exists("log.eventDataDetails")
) ||
(
  (equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
  (
    regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+add.*AppInit_DLLs") ||
    regexMatch("log.eventDataCommandLine", "(?i)reg(.exe)?\\\\s+add.*AppInit_DLLs")
  )
)
', '2026-03-02 13:28:15.154832', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1098, 'AMSI Bypass Detection', 3, 3, 2, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects attempts to bypass the Antimalware Scan Interface (AMSI) through PowerShell commands, DLL hijacking, or memory patching techniques. AMSI bypass is commonly used by attackers to evade detection when executing malicious PowerShell scripts or other code.

Next Steps:
1. Immediately isolate the affected host to prevent lateral movement
2. Examine the full PowerShell script block text in Event ID 4104 for malicious content
3. Review command line arguments in Event ID 4688 for suspicious PowerShell execution
4. Check for additional indicators of compromise on the host
5. Verify if legitimate administrative tools are being used or if this is malicious activity
6. Review recent file modifications and process execution history
7. Check for persistence mechanisms that may have been installed
8. Consider reimaging the system if compromise is confirmed
', '["https://attack.mitre.org/techniques/T1562/001/","https://docs.microsoft.com/en-us/windows/win32/amsi/antimalware-scan-interface-portal"]', e'(equals("log.eventId", "4104") &&
 ((contains("log.eventDataScriptBlockText", "[Ref].Assembly.GetType") &&
   contains("log.eventDataScriptBlockText", "amsi") &&
   contains("log.eventDataScriptBlockText", "SetValue")) ||
  contains("log.eventDataScriptBlockText", "AmsiUtils") ||
  contains("log.eventDataScriptBlockText", "amsiInitFailed") ||
  contains("log.eventDataScriptBlockText", "Bypass.AMSI") ||
  contains("log.eventDataScriptBlockText", "AmsiScanBuffer"))) ||
(equals("log.eventId", "4688") &&
 contains("log.eventDataCommandLine", "powershell") &&
 (contains("log.eventDataCommandLine", "amsi.dll") ||
  contains("log.eventDataCommandLine", "AmsiScanBuffer") ||
  contains("log.eventDataCommandLine", "amsiInitFailed")))
', '2026-03-02 13:28:16.457110', true, true, 'origin', null, '[]', '["adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1099, 'AdminSDHolder Abuse Detection', 3, 3, 2, 'Persistence, Privilege Escalation', 'T1098 - Account Manipulation', e'Detects modifications to the AdminSDHolder object which can be used for persistence by granting elevated privileges. The SDProp process propagates these permissions to protected groups every 60 minutes, making this a critical security event.

Next Steps:
1. Immediately review the user account that performed the modification
2. Check if the modification was authorized and part of legitimate administrative activities
3. Examine the specific permissions that were changed on the AdminSDHolder object
4. Monitor for privilege escalation activities in the next 60 minutes (SDProp cycle)
5. Review all members of protected groups for unauthorized additions
6. Audit recent administrative activities by the same user account
7. Consider temporarily disabling the user account if unauthorized activity is suspected
', '["https://attack.mitre.org/techniques/T1098/","https://adsecurity.org/?p=1906","https://docs.microsoft.com/en-us/windows-server/identity/ad-ds/plan/security-best-practices/appendix-c--protected-accounts-and-groups-in-active-directory"]', e'oneOf("log.eventCode", ["4662", "5136", "4670"]) &&
equals("log.channel", "Security") &&
(
  contains("log.eventDataObjectName", "CN=AdminSDHolder,CN=System") ||
  contains("log.eventDataObjectDN", "CN=AdminSDHolder,CN=System")
) &&
(
  oneOf("log.eventDataOperationType", ["Object Access", "Write Property"]) ||
  oneOf("log.eventDataAccessMask", ["0x20000", "0x40000", "0x80000"]) ||
  regexMatch("log.action", ".*Permissions.*changed.*")
) &&
!equals("log.eventDataSubjectUserName", "SYSTEM")
', '2026-03-02 13:28:17.767394', true, true, 'origin', null, '[]', '["lastEvent.log.eventDataObjectName","lastEvent.log.eventDataSubjectUserName"]');
