/* meta
time:2019-05-13 15:34:24.752698176 +0800 CST m=+0.030515698
reversion:3f71d591bce1a3ba77e74b809060cac9
down_revision:
*/

-- upgrade
CREATE TABLE test_app_00 ("name" varchar(32),"b_name" varchar(32) );
CREATE TABLE test_app_01 ("name" varchar(32),"b_name" varchar(32) );
CREATE TABLE test_app_02 ("name" varchar(32),"b_name" varchar(32) );
CREATE TABLE test_app_03 ("name" varchar(32),"b_name" varchar(32) );
CREATE TABLE test_app_04 ("name" varchar(32),"b_name" varchar(32) );
CREATE TABLE test_app_05 ("name" varchar(32),"b_name" varchar(32) );
CREATE TABLE test_app_06 ("name" varchar(32),"b_name" varchar(32) );
CREATE TABLE test_app_07 ("name" varchar(32),"b_name" varchar(32) );
CREATE TABLE test_app_08 ("name" varchar(32),"b_name" varchar(32) );
CREATE TABLE test_app_09 ("name" varchar(32),"b_name" varchar(32) )
-- end upgrade

-- downgrade
DROP TABLE test_app_00;
DROP TABLE test_app_01;
DROP TABLE test_app_02;
DROP TABLE test_app_03;
DROP TABLE test_app_04;
DROP TABLE test_app_05;
DROP TABLE test_app_06;
DROP TABLE test_app_07;
DROP TABLE test_app_08;
DROP TABLE test_app_09;
-- end downgrade