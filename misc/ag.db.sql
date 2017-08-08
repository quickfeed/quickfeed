BEGIN TRANSACTION;
CREATE TABLE "users" ("id" integer primary key autoincrement,"is_admin" bool,"name" varchar(255),"student_id" varchar(255),"email" varchar(255),"avatar_url" varchar(255) );
INSERT INTO `users` (id,is_admin,name,student_id,email,avatar_url) VALUES (1,1,'Nick Nicksen','1234','nick@example.com','https://avatars3.githubusercontent.com/u/1964338?v=4'),
 (2,0,'Bob Bobsen','1235','bob@example.com','https://avatars3.githubusercontent.com/u/1964338?v=4'),
 (3,0,'Per Pettersen','1236','per@example.com',NULL),
 (4,0,'Test Testersen','1237','test@example.com',NULL),
 (5,0,'Ola Nordmann','1238','ola@example.com',NULL),
 (6,0,'Kari Nordmann','1239','kari@example.com',NULL);
CREATE TABLE "submissions" ("id" integer primary key autoincrement,"assignment_id" bigint,"user_id" bigint,"group_id" bigint,"score" integer,"score_objects" text,"build_info" text,"commit_hash" varchar(255) );
INSERT INTO `submissions` (id,assignment_id,user_id,group_id,score,score_objects,build_info,commit_hash) VALUES (1,1,7,0,50,'[]','{}',NULL),
 (2,1,7,0,60,'[{"name": "Test 1", "score": 3, "points": 5, "weight": 100}]','{
    "builddate": "2017-07-28", 
    "buildid": 2, 
    "buildlog": "You passed", 
    "execTime": 10
}',NULL),
 (3,2,7,0,75,'[{"name": "Test 1", "score": 3, "points": 4, "weight": 100}]','{
    "builddate": "2017-07-28", 
    "buildid": 3, 
    "buildlog": "Another test", 
    "execTime": 20
}',NULL),
 (4,1,2,0,50,'[{"name": "Test 1", "score": 3, "points": 5, "weight": 100}]','{
    "builddate": "2017-07-28", 
    "buildid": 2, 
    "buildlog": "You passed", 
    "execTime": 10
}',NULL),
 (5,1,3,0,30,'[{"name": "Test 1", "score": 3, "points": 5, "weight": 100}]','{
    "builddate": "2017-07-28", 
    "buildid": 2, 
    "buildlog": "You passed", 
    "execTime": 10
}',NULL),
 (6,1,4,0,80,'[{"name": "Test 1", "score": 3, "points": 5, "weight": 100}]','{
    "builddate": "2017-07-28", 
    "buildid": 2, 
    "buildlog": "You passed", 
    "execTime": 10
}',NULL),
 (7,1,5,0,90,'[{"name": "Test 1", "score": 3, "points": 5, "weight": 100}]','{
    "builddate": "2017-07-28", 
    "buildid": 2, 
    "buildlog": "You passed", 
    "execTime": 10
}',NULL),
 (8,1,6,0,100,'[{"name": "Test 1", "score": 3, "points": 5, "weight": 100}]','{
    "builddate": "2017-07-28", 
    "buildid": 2, 
    "buildlog": "You passed", 
    "execTime": 10
}',NULL),
 (9,3,1,0,60,'[{"name": "Test 1", "score": 3, "points": 5, "weight": 100}]','{
    "builddate": "2017-07-28", 
    "buildid": 2, 
    "buildlog": "You passed", 
    "execTime": 10
}',NULL);
CREATE TABLE "remote_identities" ("id" integer primary key autoincrement,"provider" varchar(255),"remote_id" bigint,"access_token" varchar(255),"user_id" bigint );
CREATE TABLE "groups" ("id" integer primary key autoincrement,"name" varchar(255),"status" integer,"course_id" bigint );
CREATE TABLE "enrollments" ("id" integer primary key autoincrement,"course_id" bigint,"user_id" bigint,"group_id" bigint,"status" integer );
INSERT INTO `enrollments` (id,course_id,user_id,group_id,status) VALUES (1,1,1,0,2),
 (2,2,1,0,2),
 (3,1,2,0,0),
 (4,2,2,0,2),
 (5,4,2,0,0),
 (6,2,3,0,2),
 (7,4,3,0,0),
 (8,5,4,0,2),
 (9,3,4,0,0),
 (10,5,5,0,2),
 (11,3,5,0,0),
 (12,4,5,0,2),
 (13,5,6,0,2),
 (14,3,6,0,0),
 (15,1,3,0,2),
 (16,1,4,0,2),
 (17,1,5,0,0),
 (18,1,6,0,0);
CREATE TABLE "courses" ("id" integer primary key autoincrement,"name" varchar(255),"code" varchar(255),"year" integer,"tag" varchar(255),"provider" varchar(255),"directory_id" bigint );
INSERT INTO `courses` (id,name,code,year,tag,provider,directory_id) VALUES (1,'TST100','TST100',2017,'fall','fake',123122),
 (2,'Objectoriented programming','DAT100',2018,'spring','fake',123123),
 (3,'Operatingsystem','DAT320',2017,'fall','fake',123124),
 (4,'Algorithms and data structures','DAT200',2017,'fall','fake',123125),
 (5,'Databases','DAT220',2018,'sprint','fake',123126);
CREATE TABLE "assignments" ("id" integer primary key autoincrement,"course_id" bigint,"name" varchar(255),"language" varchar(255),"deadline" datetime,"auto_approve" bool DEFAULT false,"order" integer );
INSERT INTO `assignments` (id,course_id,name,"language",deadline,auto_approve,"order") VALUES (1,1,'test','java','2017-07-29T22:00:00Z','false',1),
 (2,1,'test2','java','2017-07-30T22:00:00Z','false',2),
 (3,2,'Output','java','2017-09-15T00:00:00Z','false',1),
 (4,2,'Variables','java','2017-09-20T00:00:00Z','false',2),
 (5,3,'Learn GO','go','2017-09-15T00:00:00Z','false',1),
 (6,3,'Goroutine','go','2017-09-20T00:00:00Z','false',2),
 (7,4,'Binary Tree','java','2017-09-15T00:00:00Z','false',1),
 (8,4,'Hashmap','java','2017-09-20T00:00:00Z','false',2),
 (9,5,'SELECT','sql','2017-09-15T00:00:00Z','false',1),
 (10,5,'UPDATE','sql','2017-09-20T00:00:00Z','false',2);
CREATE UNIQUE INDEX idx_unique_group_name ON "groups"("name", course_id) ;
COMMIT;
