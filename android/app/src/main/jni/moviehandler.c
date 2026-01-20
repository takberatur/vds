#include <jni.h>
#include <string.h>
#include <stdlib.h>


const char *const JNIREG_CLASS = "com/agcforge/lk21xxi/videodownloader/utils/AppManager";


//packagename
static const char *PKG = "com.agcforge.lk21xxi.videodownloader";

//tapdaq
const char *const appid = "";
const char *const clientkey = "";

//FAN
const char *const fanid = "";

jobject getApplication(JNIEnv *env) {
    jobject application = NULL;

    jclass activity_thread_clz = (*env)->FindClass(env,"android/app/ActivityThread");
    if (activity_thread_clz != NULL) {
        jmethodID currentApplication = (*env)->GetStaticMethodID(env,
                                                                 activity_thread_clz, "currentApplication", "()Landroid/app/Application;");
        if (currentApplication != NULL) {
            application = (*env)->CallStaticObjectMethod(env,activity_thread_clz, currentApplication);
        } else {
            //LOGE("Cannot find method: currentApplication() in ActivityThread.");
        }
        (*env)->DeleteLocalRef(env,activity_thread_clz);
    } else {
        //LOGE("Cannot find class: android.app.ActivityThread");
    }

    return application;
}

int checkSign(JNIEnv *env) {

    jobject application = getApplication(env);
    if (application == NULL) {
        return JNI_ERR;
    }
    // Context(ContextWrapper) class
    jclass context_clz = (*env)->GetObjectClass(env,application);
    // getPackageManager()
    jmethodID getPackageManager = (*env)->GetMethodID(env,context_clz, "getPackageManager",
                                                      "()Landroid/content/pm/PackageManager;");
    // android.content.pm.PackageManager object
    jobject package_manager = (*env)->CallObjectMethod(env,application, getPackageManager);
    // PackageManager class
    jclass package_manager_clz = (*env)->GetObjectClass(env,package_manager);
    // getPackageInfo()
    jmethodID getPackageInfo = (*env)->GetMethodID(env,package_manager_clz, "getPackageInfo",
                                                   "(Ljava/lang/String;I)Landroid/content/pm/PackageInfo;");
    // context.getPackageName()
    jmethodID getPackageName = (*env)->GetMethodID(env,context_clz, "getPackageName",
                                                   "()Ljava/lang/String;");
    // call getPackageName() and cast from jobject to jstring
    jstring package_name = (jstring) ((*env)->CallObjectMethod(env,application, getPackageName));

    // field signatures


    // release
    (*env)->DeleteLocalRef(env,application);
    (*env)->DeleteLocalRef(env,context_clz);
    (*env)->DeleteLocalRef(env,package_manager);
    (*env)->DeleteLocalRef(env,package_manager_clz);


    const char *pkg = (*env)->GetStringUTFChars(env,package_name, NULL);

    int result2 = strcmp(pkg, PKG);

//#ifdef DEBUG
    //   __android_log_print(ANDROID_LOG_ERROR, "maxgba", "%s ",sign);
//
//#endif

    // 使用之后要释放这段内存

//
    (*env)->ReleaseStringUTFChars(env,package_name, pkg);
    (*env)->DeleteLocalRef(env,package_name);


    if (result2 == 0) { // 签名一致
        return JNI_OK;
    }

    return JNI_ERR;
}

void verifySign(JNIEnv *env) {

    if (checkSign(env) == JNI_ERR) {
        exit(1);
    }
}

static jobject getGlobalContext(JNIEnv *env)
{

    jclass activityThread = (*env)->FindClass(env,"android/app/ActivityThread");
    jmethodID currentActivityThread = (*env)->GetStaticMethodID(env,activityThread, "currentActivityThread", "()Landroid/app/ActivityThread;");
    jobject at = (*env)->CallStaticObjectMethod(env,activityThread, currentActivityThread);

    jmethodID getApplication = (*env)->GetMethodID(env,activityThread, "getApplication", "()Landroid/app/Application;");
    jobject context = (*env)->CallObjectMethod(env,at, getApplication);
    return context;
}

JNIEXPORT jstring JNICALL
initLib(JNIEnv *env, jobject thiz) {
    return (*env)->NewStringUTF(env, "init rom, don't remove");
    //init(env, thiz);
}

JNIEXPORT jstring JNICALL
fan(JNIEnv *env, jobject thiz) {
    return (*env)->NewStringUTF(env, fanid);
    //init(env, thiz);
}

JNIEXPORT void JNICALL
initAdsNetwork(JNIEnv *env, jobject thiz, jobject activity) {

jstring appidX = (*env)->NewStringUTF(env,appid);
jstring clientkeyX = (*env)->NewStringUTF(env,clientkey);


jobject listener = (*env)->CallStaticObjectMethod(env, app_clz, getlistener);

(*env)->CallVoidMethod(env, tapdaqInstance, funcInit, activity,appidX,clientkeyX,NULL, listener);

}



JNINativeMethod method_table[] = {
        // {"enabled", "(Ljava/lang/Boolean;)V", (void *) setEnabledAds},
        {"init", "()Ljava/lang/String;", (void *) initLib},
        {"getId", "()Ljava/lang/String;", (void *) fan},
        {"register","(Landroid/app/Activity;)V", (void *) initAdsNetwork},

        //绑定
};


int registerNativeMethods(JNIEnv *env) {
    jclass clazz;
    clazz = (*env)->FindClass(env,JNIREG_CLASS);
    if (clazz == NULL) {
        return JNI_FALSE;
    }
    if ((*env)->RegisterNatives(env,clazz, method_table, sizeof(method_table) / sizeof(method_table[0])) <
        0) {
        return JNI_FALSE;
    }

    return JNI_TRUE;
}


JNIEXPORT jint JNICALL JNI_OnLoad(JavaVM *vm, void *reserved) {
    JNIEnv *env = NULL;
    // jobject obj = getApplication(env);

    if ((*vm)->GetEnv(vm, (void **) &env, JNI_VERSION_1_6) != JNI_OK){
        //LOGE("Failed to get the environment");
        return JNI_ERR;
    }
    verifySign(env);

    if (registerNativeMethods(env) == JNI_FALSE) {
        exit(0);
    };

//    if (checkSign(env) == JNI_ERR) {
//        exit(0);
//
//    }

    return JNI_VERSION_1_6;
}






